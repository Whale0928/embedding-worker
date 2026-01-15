package cmd

import (
	"fmt"
	"os"

	"github.com/Whale0928/embedding-worker/internal/downloader"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "다운로드 받은 모델 파일을 삭제한다.",
	RunE:  runCleanUp,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func runCleanUp(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Embedder Worker - Clean Up ===")
	fmt.Println()

	cfg := GetConfig()

	fmt.Printf("[1] 캐시 디렉토리: %s\n", cfg.CacheDir)
	fmt.Println()

	fmt.Println("[2] 모델 파일 삭제...")
	deletedCount := 0

	for _, filename := range downloader.ModelFiles {

		filePath := cfg.CacheDir + "/" + filename

		if _, err := os.Stat(filePath); err == nil {
			if err := os.Remove(filePath); err == nil {
				fmt.Printf("    [삭제] %s\n", filename)
				deletedCount++
			} else {
				fmt.Printf("    [실패] %s: %v\n", filename, err)
			}
		} else {
			fmt.Printf("    [없음] %s\n", filename)
		}
	}
	fmt.Println()

	fmt.Printf("=== Clean Up 완료: %d개 파일 삭제 ===\n", deletedCount)
	return nil
}
