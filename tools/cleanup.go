package mtp

import (
	"fmt"
	"os"
	"path/filepath"
)

func CleanupTempFolder(folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		target := filepath.Join(folder, entry.Name())
		if removeErr := os.RemoveAll(target); removeErr != nil {
			return fmt.Errorf("failed cleaning %q: %w", target, removeErr)
		}
	}

	return nil
}
