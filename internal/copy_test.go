package copyFiles

import (
	mtp "Velox/tools"
	"context"
	"log"
	"testing"
)

func BenchmarkBulkCopy(b *testing.B) {

	b.Run("Real Data Test", func(b *testing.B) {
		ctx := context.Background()
		destDir := "/home/yeray/Pictures/Temp"
		sourceDir, files, err := GetMTPCameraFile(ctx, "jpg")
		if err != nil {
			log.Printf("Error getting mobile path and files: %v", err.Error())
		}
		for i := 0; i < b.N; i++ {
			if BulkCopy(ctx, sourceDir, files, destDir, mtp.JoinMTP, mtp.CopyFromMTP).Error() != "" {
				log.Printf("BulkCopy Error during BenchmarkBulkCopy Real Data Test: %v",
					BulkCopy(ctx, sourceDir, files, destDir, mtp.JoinMTP, mtp.CopyFromMTP).Error())
			}
		}
	})
	b.Run("Mock Data Test", func(b *testing.B) {
		ctx := context.Background()
		destDir := "/home/yeray/Pictures/VeloxMockTemp"
		sourceDir, files, err := GetMTPCameraFile(ctx, "jpg")
		if err != nil {
			log.Printf("Error getting mobile path and files: %v", err.Error())
		}
		for i := 0; i < b.N; i++ {
			if BulkCopy(ctx, sourceDir, files, destDir, LocalJoin, CopyFromTmpFolder).Error() != "" {
				log.Printf("BulkCopy Error during BenchmarkBulkCopy Mock Data Test: %v",
					BulkCopy(ctx, sourceDir, files, destDir, LocalJoin, CopyFromTmpFolder).Error())
			}
		}
	})
}
