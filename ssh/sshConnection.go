package ssh

import (
	"Velox/immich"
	copyFiles "Velox/internal"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func SSHConnection(maxConcurrency int64, ctx context.Context, sourceDir, destDir string, files []string, join copyFiles.JoinFunc) error {
	user := os.Getenv("VM_USER")
	ip := os.Getenv("VM_IP")
	auth := os.Getenv("VM_AUTH")
	immichURL := os.Getenv("IMMICH_INSTANCE_URL")
	immichAPIKey := os.Getenv("IMMICH_API_KEY")

	if immichURL == "" || immichAPIKey == "" {
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

	sftpClient, err := sftp.NewClient(conn)
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
	return nil
}

func copyBatchViaSFTP(ctx context.Context, sftpClient *sftp.Client, src []string, dst string) error {
	if err := sftpClient.MkdirAll(dst); err != nil {
		return err
	}

	for _, srcPath := range src {
		fileName := filepath.Base(srcPath)
		dstPath := filepath.Join(dst, fileName)
		if err := copyOneViaSFTP(ctx, sftpClient, srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
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

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}
