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
)

var (
	errPhoneMountNotFound = errors.New("phone mtp mount not found")
	mtpRootRe             = regexp.MustCompile(`default_location=(mtp://[^/]+/)`)
)

func DetectPhoneMountRootURI(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gio", "mount", "-li")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gio mount -li failed: %w\n%s", err, string(out))
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
	cmd := exec.CommandContext(ctx, "gio", "list", dirURI)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gio list failed: %w\n%s", err, string(out))
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

	cmd := exec.CommandContext(ctx, "gio", "copy", "--", srcURI, dstURI)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gio copy failed: %w\n%s", err, string(out))
	}
	return nil
}

func JoinMTP(baseURI, fileName string) string {
	baseURI = strings.TrimRight(baseURI, "/")
	return baseURI + "/" + url.PathEscape(fileName)
}
