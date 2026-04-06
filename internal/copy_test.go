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
			b.Fatalf("Error getting mobile path and files: %v", err.Error())
		}
		for i := 0; i < b.N; i++ {
			if BulkCopy(ctx, sourceDir, files, destDir, mtp.JoinMTP, mtp.CopyFromMTP) != nil {
				b.Fatal("BulkCopy Error during BenchmarkBulkCopy Real Data Test")
			}
		}
	})
	b.Run("Mock Data Test", func(b *testing.B) {
		ctx := context.Background()
		destDir := "/home/yeray/Pictures/VeloxMockTemp"
		sourceDir := "/home/yeray/Pictures/VeloxMock"
		files, err := ListAllFiles(sourceDir, "jpg")
		if err != nil {
			log.Printf("Error getting mobile path and files: %v", err.Error())
		}
		for i := 0; i < b.N; i++ {
			if BulkCopy(ctx, sourceDir, files, destDir, LocalJoin, CopyFromTmpFolder) != nil {
				b.Fatal("BulkCopy Error during BenchmarkBulkCopy Mock Data Test")
			}
		}
	})
}
