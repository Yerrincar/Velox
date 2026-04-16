package ssh

import (
	"Velox/immich"
	copyFiles "Velox/internal"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftpJob struct {
	src []string
	dst string
}

func SSHConnection(maxConcurrency int64, ctx context.Context, sourceDir, destDir string, files []string, join copyFiles.JoinFunc, runImmichUpload bool) error {
	if len(files) == 0 {
		return nil
	}
	jobs := make(chan sftpJob)
	errCh := make(chan error, maxConcurrency)
	var wg sync.WaitGroup
	workerCount := int(maxConcurrency)
	if workerCount < 1 {
		workerCount = 1
	}

	for i := 0; i < workerCount; i++ {
		wg.Go(func() {
			conn, sftpClient, err := openSFTPClient()
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()
			defer sftpClient.Close()
			for job := range jobs {
				if err := copyBatchViaSFTP(ctx, sftpClient, job.src, job.dst); err != nil {
					errCh <- err
					return
				}
			}
		})
	}
	for chunk := range slices.Chunk(files, 50) {
		srcs := make([]string, 0, len(chunk))
		for _, f := range chunk {
			srcs = append(srcs, join(sourceDir, f))
		}
		select {
		case jobs <- sftpJob{src: srcs, dst: destDir}:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	if runImmichUpload {
		return runRemoteImmichUpload(ctx, destDir)
	}
	return nil
}

func openSFTPClient() (*ssh.Client, *sftp.Client, error) {
	user := os.Getenv("VM_USER")
	ip := os.Getenv("VM_IP")
	auth := os.Getenv("VM_AUTH")

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            user,
		Auth: []ssh.AuthMethod{
			ssh.Password(auth),
		},
		Timeout: 10 * time.Second,
	}
	conn, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), config)
	if err != nil {
		return nil, nil, err
	}

	sftpClient, err := sftp.NewClient(
		conn,
		sftp.UseConcurrentWrites(true),
		sftp.MaxConcurrentRequestsPerFile(128),
	)
	if err != nil {
		return nil, nil, err
	}

	return conn, sftpClient, nil
}

func runRemoteImmichUpload(ctx context.Context, destDir string) error {
	conn, _, err := openSFTPClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	immichURL := os.Getenv("IMMICH_INSTANCE_URL")
	immichAPIKey := os.Getenv("IMMICH_API_KEY")
	if immichURL == "" || immichAPIKey == "" {
		return errors.New("IMMICH_INSTANCE_URL and IMMICH_API_KEY must be set")
	}
	cmd := immich.ImmichUpload(destDir, immichURL, immichAPIKey)
	done := make(chan error, 1)
	go func() {
		done <- session.Run(cmd)
	}()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("remote immich upload failed:  %w\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
		}
		return nil
	case <-ctx.Done():
		_ = session.Close()
		return ctx.Err()
	}
}

func copyBatchViaSFTP(ctx context.Context, sftpClient *sftp.Client, src []string, dst string) error {
	const perBatchConcurrency = 4

	if err := sftpClient.MkdirAll(dst); err != nil {
		return err
	}

	workerCount := min(perBatchConcurrency, len(src))
	sem := make(chan struct{}, workerCount)
	errChan := make(chan error, len(src))
	var wg sync.WaitGroup

	for _, srcPath := range src {
		if err := ctx.Err(); err != nil {
			return err
		}

		srcPath := srcPath
		fileName := filepath.Base(srcPath)
		dstPath := filepath.Join(dst, fileName)

		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()

			if err := copyOneViaSFTP(ctx, sftpClient, srcPath, dstPath); err != nil {
				errChan <- err
			}
		})
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var firstErr error
	for err := range errChan {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func copyOneViaSFTP(ctx context.Context, sftpClient *sftp.Client, src, dst string) error {
	_ = ctx

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := sftpClient.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = destinationFile.ReadFrom(sourceFile)
	if err != nil {
		_ = sftpClient.Remove(dst)
	}
	return err
}
