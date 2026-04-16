package mtp

import (
	"os"
	"path/filepath"
)

func DefaultLocalStagingDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		return filepath.Join(os.TempDir(), "velox", "staging")
	}

	return filepath.Join(cacheDir, "velox", "staging")
}
