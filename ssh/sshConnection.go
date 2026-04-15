package ssh

import (
	"Velox/immich"
	copyFiles "Velox/internal"
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func SSHConnection(maxConcurrency int64, ctx context.Context, sourceDir, destDir string, files []string, join copyFiles.JoinFunc, runImmichUpload bool) error {
	user := os.Getenv("VM_USER")
	ip := os.Getenv("VM_IP")
	auth := os.Getenv("VM_AUTH")
	immichURL := os.Getenv("IMMICH_INSTANCE_URL")
	immichAPIKey := os.Getenv("IMMICH_API_KEY")

	if runImmichUpload && (immichURL == "" || immichAPIKey == "") {
		return errors.New("IMMICH_INSTANCE_URL and IMMICH_API_KEY must be set")
	}

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
		return err
	}
	defer conn.Close()

	sftpClient, err := sftp.NewClient(
		conn,
		sftp.UseConcurrentWrites(true),
		sftp.MaxConcurrentRequestsPerFile(128),
	)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	copyViaSFTP := func(copyCtx context.Context, src []string, dst string) error {
		return copyBatchViaSFTP(copyCtx, sftpClient, src, dst)
	}

	if err := copyFiles.BulkCopy(maxConcurrency, ctx, sourceDir, files, destDir, join, copyViaSFTP); err != nil {
		return err
	}

	if runImmichUpload {
		session, err := conn.NewSession()
		if err != nil {
			return err
		}
		defer session.Close()

		cmdCommand := immich.ImmichUpload(destDir, immichURL, immichAPIKey)
		err = session.Run(cmdCommand)
		if err != nil {
			return err
		}
	}
	return nil
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
