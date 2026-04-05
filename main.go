package main

import (
	copyFiles "Velox/internal"
	mtp "Velox/tools"
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	destDir := "/home/yeray/Pictures/Temp/"
	sourceDir, files, err := copyFiles.GetMTPCameraFile(ctx, "jpg")
	if err != nil {
		log.Print(err.Error())
	}

	err = copyFiles.BulkCopy(ctx, sourceDir, files, destDir, mtp.JoinMTP, mtp.CopyFromMTP)
	if err != nil {
		log.Print(err.Error())
	}
}
