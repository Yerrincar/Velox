package main

import (
	copyFiles "Velox/internal"
	mtp "Velox/tools"
	"context"
	"log"
	"path/filepath"
	"time"
)

func main() {
	setupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	destDir, err := filepath.Abs("/home/yeray/Pictures/Temp/")
	if err != nil {
		log.Print(err.Error())
	}
	sourceDir, files, err := copyFiles.GetMTPCameraFile(setupCtx, "jpg")
	if err != nil {
		log.Print(err.Error())
	}

	runCtx := context.Background()
	err = copyFiles.BulkCopy(4, runCtx, sourceDir, files, destDir, mtp.JoinMTP, mtp.CopyFromMTP)
	if err != nil {
		log.Print(err.Error())
	}
}
