package mtp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	errPhoneMountNotFound = errors.New("phone mtp mount not found")
	mtpRootRe             = regexp.MustCompile(`default_location=(mtp://[^/]+/)`)
)

func DetectPhoneMountRootURI(ctx context.Context) (string, error) {
	var out []byte
	err := reusableRetries(ctx, 3, func() error {
		cmd := exec.CommandContext(ctx, "gio", "mount", "-li")
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
		cmd := exec.CommandContext(ctx, "gio", "list", dirURI)
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

func CopyFromMTP(ctx context.Context, srcURI, localDstPath string) error {
	absDst, err := filepath.Abs(localDstPath)
	if err != nil {
		return err
	}
	dstURI := (&url.URL{Scheme: "file", Path: absDst}).String()

	return reusableRetries(ctx, 3, func() error {
		cmd := exec.CommandContext(ctx, "gio", "copy", "--", srcURI, dstURI)
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
