package copyFiles

import (
	"Velox/tools"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type JoinFunc func(base, file string) string
type CopyFunc func(ctx context.Context, src []string, dst string) error

type Semaphore struct {
	Channel chan struct{}
}

func (s *Semaphore) Acquire() {
	s.Channel <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.Channel
}

func BulkCopy(maxConcurrency int64, ctx context.Context, sourceDir string, files []string, destDir string, join JoinFunc, copyOneChunk CopyFunc) error {
	const perBatchTimeout = 3 * time.Minute
	chunkSize := 200
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}
	count := (len(files) + chunkSize - 1) / chunkSize
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	sem := &Semaphore{Channel: make(chan struct{}, maxConcurrency)}
	var wg sync.WaitGroup
	errChan := make(chan error, count)
	for chunk := range slices.Chunk(files, chunkSize) {
		srcs := make([]string, 0, len(chunk))
		for _, f := range chunk {
			srcs = append(srcs, join(sourceDir, f))
		}
		batch := srcs
		sem.Acquire()
		wg.Go(func() {
			defer sem.Release()
			copyCtx, cancel := context.WithTimeout(ctx, perBatchTimeout)
			defer cancel()
			err := copyOneChunk(copyCtx, batch, absDestDir)
			if err != nil {
				errChan <- err
			}
		})
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var firstError error
	for err := range errChan {
		if err != nil && firstError == nil {
			firstError = err
		}
	}
	return firstError
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
	targetedFiles := make([]string, 0, len(allFiles))
	for _, f := range allFiles {
		if strings.HasSuffix(f, suffix) {
			targetedFiles = append(targetedFiles, f)
		}
	}
	return filesPath, targetedFiles, nil
}

func CopyFromTmpFolder(ctx context.Context, src, dst string) error {
	_ = ctx

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func CopyBatchFromTmpFolder(ctx context.Context, src []string, dst string) error {
	for _, srcPath := range src {
		fileName := filepath.Base(srcPath)
		dstPath := filepath.Join(dst, fileName)
		if err := CopyFromTmpFolder(ctx, srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func LocalJoin(base, file string) string {
	return filepath.Join(base, file)
}

func ListAllFiles(src, suffix string) ([]string, error) {
	sourceFolder, err := os.ReadDir(src)
	if err != nil {
		log.Printf("Error trying to open source mock data folder: %v", err.Error())
		return nil, err
	}
	files := make([]string, 0, len(sourceFolder))
	for _, file := range sourceFolder {
		if strings.HasSuffix(file.Name(), suffix) {
			files = append(files, file.Name())
		}
	}

	return files, nil
}
