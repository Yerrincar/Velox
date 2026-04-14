package main

import (
	copyFiles "Velox/internal"
	"Velox/ssh"
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
	sourceDir, files, err := copyFiles.GetADBCameraFile(setupCtx, "jpg")
	if err != nil {
		log.Print(err.Error())
		return
	}

	runCtx := context.Background()
	err = copyFiles.BulkCopy(3, runCtx, sourceDir, files, destDir, mtp.JoinADB, mtp.CopyFromADB)
	if err != nil {
		log.Print(err.Error())
		return
	}

	sourceTempDir := destDir
	tempFiles, err := copyFiles.ListAllFiles(sourceTempDir, "jpg")
	if err != nil {
		log.Print(err.Error())
		return
	}
	vmDestDir := "/var/tmp/velox-staging"
	err = ssh.SSHConnection(5, runCtx, sourceTempDir, vmDestDir, tempFiles, copyFiles.LocalJoin)
	if err != nil {
		log.Print(err.Error())
		return
	}

	err = mtp.CleanupTempFolder(sourceTempDir)
	if err != nil {
		log.Print(err.Error())
		return
	}
}
