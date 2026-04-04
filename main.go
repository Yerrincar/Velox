package main

import (
	"Velox/tools"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	err := bulkCopy()
	if err != nil {
		log.Print(err.Error())
	}
}

func bulkCopy() error {
	ctx := context.Background()
	rootURI, err := mtp.DetectPhoneMountRootURI(ctx)
	if err != nil {
		return err
	}

	filesPath := strings.TrimRight(rootURI, "/") + "/Almacenamiento interno compartido/DCIM/Camera/"
	path, err := mtp.ListMTPFiles(ctx, filesPath)
	if err != nil {
		return err
	}
	for _, f := range path {
		if !strings.HasSuffix(f, "jpg") {
			continue
		}
		err := mtp.CopyFromMTP(ctx, mtp.JoinMTP(filesPath, f), filepath.Join("/home/yeray/Pictures/Temp", f))
		if err != nil {
			return err
		}
	}
	return nil
}

func copySingleFile(src, dst string) error {
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
