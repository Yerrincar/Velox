package mtp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	errPhoneMountNotFound = errors.New("phone mtp mount not found")
	mtpRootRe             = regexp.MustCompile(`default_location=(mtp://[^/]+/)`)
	gioPath               = "/usr/bin/gio"
	adbPath               = "/usr/bin/adb"
)

func DetectPhoneMountRootURI(ctx context.Context) (string, error) {
	var out []byte
	err := reusableRetries(ctx, 3, func() error {
		cmd := exec.CommandContext(ctx, gioPath, "mount", "-li")
		cmdOut, cmdErr := cmd.CombinedOutput()
		out = cmdOut
		if cmdErr != nil {
			return fmt.Errorf("gio mount -li failed: %w\n%s", cmdErr, string(cmdOut))
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		m := mtpRootRe.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}
		return m[1], nil
	}

	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("reading gio output: %w", err)
	}
	return "", errPhoneMountNotFound
}

func ListMTPFiles(ctx context.Context, dirURI string) ([]string, error) {
	var out []byte
	err := reusableRetries(ctx, 3, func() error {
		cmd := exec.CommandContext(ctx, gioPath, "list", dirURI)
		cmdOut, cmdErr := cmd.CombinedOutput()
		out = cmdOut
		if cmdErr != nil {
			return fmt.Errorf("gio list failed: %w\n%s", cmdErr, string(cmdOut))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	files := make([]string, 0, len(lines))
	for _, l := range lines {
		name := strings.TrimSpace(l)
		if name == "" || strings.HasSuffix(name, "/") {
			continue
		}
		files = append(files, name)
	}
	return files, nil
}

func CopyFromMTP(ctx context.Context, srcURI []string, localDstPath string) error {
	dstURI := (&url.URL{Scheme: "file", Path: localDstPath}).String()

	args := []string{"copy", "--"}
	args = append(append(args, srcURI...), dstURI)
	return reusableRetries(ctx, 3, func() error {
		cmd := exec.CommandContext(ctx, gioPath, args...)
		out, cmdErr := cmd.CombinedOutput()
		if cmdErr != nil {
			return fmt.Errorf("gio copy failed: %w\n%s", cmdErr, string(out))
		}
		return nil
	})
}

func JoinMTP(baseURI, fileName string) string {
	baseURI = strings.TrimRight(baseURI, "/")
	return baseURI + "/" + url.PathEscape(fileName)
}

func JoinADB(basePath, fileName string) string {
	basePath = strings.TrimRight(basePath, "/")
	return basePath + "/" + fileName
}

func EnsureADBDevice(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, adbPath, "devices")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb devices failed: %w\n%s", err, string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.HasSuffix(strings.TrimSpace(line), "\tdevice") {
			return nil
		}
	}

	return errors.New("no authorized adb device found")
}

func ListADBFiles(ctx context.Context, remoteDir string) ([]string, error) {
	cmd := exec.CommandContext(ctx, adbPath, "shell", "ls", "-1", remoteDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("adb shell ls failed: %w\n%s", err, string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	files := make([]string, 0, len(lines))
	for _, l := range lines {
		name := strings.TrimSpace(strings.TrimSuffix(l, "\r"))
		if name == "" || strings.HasSuffix(name, "/") {
			continue
		}
		files = append(files, name)
	}
	return files, nil
}

func CopyFromADB(ctx context.Context, src []string, localDstDir string) error {
	if err := os.MkdirAll(localDstDir, 0o755); err != nil {
		return err
	}

	for _, remotePath := range src {
		localPath := filepath.Join(localDstDir, path.Base(remotePath))
		cmd := exec.CommandContext(ctx, adbPath, "pull", "-a", remotePath, localPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("adb pull failed: %w\n%s", err, string(out))
		}
	}

	return nil
}

func reusableRetries(ctx context.Context, maxRetries int, f func() error) error {
	if maxRetries < 1 {
		maxRetries = 1
	}

	var err error
	for i := 0; i < maxRetries; i++ {
		if err = f(); err == nil {
			return nil
		}

		if i == maxRetries-1 {
			break
		}

		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}
