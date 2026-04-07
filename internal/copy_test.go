package copyFiles

import (
	mtp "Velox/tools"
	"context"
	"testing"
	"time"
)

func BenchmarkBulkCopy(b *testing.B) {

	b.Run("Real Data Test", func(b *testing.B) {
		setupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		destDir := "/home/yeray/Pictures/Temp"
		sourceDir, files, err := GetMTPCameraFile(setupCtx, "jpg")
		if err != nil {
			b.Fatalf("Error getting mobile path and files: %v", err.Error())
		}

		runCtx := context.Background()
		for i := 0; i < b.N; i++ {
			if BulkCopy(3, runCtx, sourceDir, files, destDir, mtp.JoinMTP, mtp.CopyFromMTP) != nil {
				b.Fatalf("BulkCopy Error during BenchmarkBulkCopy Real Data Test")
			}
		}
	})
	b.Run("Mock Data Test", func(b *testing.B) {
		ctx := context.Background()
		destDir := "/home/yeray/Pictures/VeloxMockTemp"
		sourceDir := "/home/yeray/Pictures/VeloxMock"
		files, err := ListAllFiles(sourceDir, "jpg")
		if err != nil {
			b.Fatalf("Error getting local path and files: %v", err.Error())
		}
		for i := 0; i < b.N; i++ {
			if BulkCopy(3, ctx, sourceDir, files, destDir, LocalJoin, CopyFromTmpFolder) != nil {
				b.Fatalf("BulkCopy Error during BenchmarkBulkCopy Mock Data Test")
			}
		}
	})
}
