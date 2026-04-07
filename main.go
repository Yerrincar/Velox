package main

import (
	copyFiles "Velox/internal"
	mtp "Velox/tools"
	"context"
	"log"
	"time"
)

func main() {
	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	defer cancel()
	destDir := "/home/yeray/Pictures/Temp/"
	sourceDir, files, err := copyFiles.GetMTPCameraFile(ctx, "jpg")
	if err != nil {
		log.Print(err.Error())
	}

	err = copyFiles.BulkCopy(3, ctxTimeout, sourceDir, files, destDir, mtp.JoinMTP, mtp.CopyFromMTP)
	if err != nil {
		log.Print(err.Error())
	}
}
