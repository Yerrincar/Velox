package copyFiles

import (
	"Velox/tools"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type JoinFunc func(base, file string) string
type CopyFunc func(ctx context.Context, src, dst string) error

func BulkCopy(ctx context.Context, sourceDir string, files []string, destDir string, join JoinFunc, copyOne CopyFunc) error {
	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			src := join(sourceDir, fileName)
			dst := filepath.Join(destDir, fileName)

			if err := copyOne(ctx, src, dst); err != nil {
				log.Printf("Failed to copy %s: %v", fileName, err)
			}
		}(f)
	}
	wg.Wait()
	return nil
}

func GetMTPCameraFile(ctx context.Context, suffix string) (string, []string, error) {
	rootURI, err := mtp.DetectPhoneMountRootURI(ctx)
	if err != nil {
		return "", nil, err
	}

	filesPath := strings.TrimRight(rootURI, "/") + "/Almacenamiento interno compartido/DCIM/Camera/"
	allFiles, err := mtp.ListMTPFiles(ctx, filesPath)
	if err != nil {
		return "", nil, err
	}
	targetedFiles := make([]string, len(allFiles))
	for _, f := range allFiles {
		if strings.HasSuffix(f, suffix) {
			targetedFiles = append(targetedFiles, f)
		}
	}
	return filesPath, targetedFiles, nil
}

func CopyFromTmpFolder(ctx context.Context, src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	fmt.Println("File copied successfully")
	return nil
}

func LocalJoin(base, file string) string {
	return filepath.Join(base, file)
}
