package main

import (
	copyFiles "Velox/internal"
	"Velox/ssh"
	mtp "Velox/tools"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, using system environment variables")
	}

	transfer := flag.String("transfer", "full", "Transfer scope: full|semi|partial")
	mode := flag.String("mode", "adb", "Mobile mode: adb|mtp (recommended: adb)")
	vmIP := flag.String("ip", "", "VM IP override (optional, overrides VM_IP)")
	vmFolder := flag.String("folder", "", "VM destination folder (default: /var/tmp/velox-staging)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Velox - photo transfer pipeline\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  velox [--transfer full|semi|partial] [--mode adb|mtp] [--ip <vm-ip>] [--folder <vm-folder>]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	*transfer = strings.ToLower(strings.TrimSpace(*transfer))
	*mode = strings.ToLower(strings.TrimSpace(*mode))

	if *transfer != "full" && *transfer != "semi" && *transfer != "partial" {
		log.Print("invalid --transfer value, use: full|semi|partial")
		flag.Usage()
		return
	}

	if *mode != "adb" && *mode != "mtp" {
		log.Print("invalid --mode value, use: adb|mtp")
		flag.Usage()
		return
	}

	if *vmIP != "" {
		if err := os.Setenv("VM_IP", *vmIP); err != nil {
			log.Print(err.Error())
			return
		}
	}

	setupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	destDir, err := filepath.Abs("/home/yeray/Pictures/Temp/")
	if err != nil {
		log.Print(err.Error())
		return
	}

	var (
		sourceDir string
		files     []string
		join      copyFiles.JoinFunc
		copyBatch copyFiles.CopyFunc
	)

	if *mode == "adb" {
		sourceDir, files, err = copyFiles.GetADBCameraFile(setupCtx, "jpg")
		join = mtp.JoinADB
		copyBatch = mtp.CopyFromADB
	} else {
		sourceDir, files, err = copyFiles.GetMTPCameraFile(setupCtx, "jpg")
		join = mtp.JoinMTP
		copyBatch = mtp.CopyFromMTP
	}
	if err != nil {
		log.Print(err.Error())
		return
	}

	runCtx := context.Background()
	err = copyFiles.BulkCopy(3, runCtx, sourceDir, files, destDir, join, copyBatch)
	if err != nil {
		log.Print(err.Error())
		return
	}

	if *transfer == "partial" {
		return
	}

	sourceTempDir := destDir
	tempFiles, err := copyFiles.ListAllFiles(sourceTempDir, "jpg")
	if err != nil {
		log.Print(err.Error())
		return
	}
	vmDestDir := strings.TrimSpace(*vmFolder)
	if vmDestDir == "" {
		vmDestDir = "/var/tmp/velox-staging"
	}

	runImmichUpload := *transfer == "full"
	err = ssh.SSHConnection(5, runCtx, sourceTempDir, vmDestDir, tempFiles, copyFiles.LocalJoin, runImmichUpload)
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
