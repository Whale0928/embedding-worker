package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Whale0928/embedding-worker/internal/downloader"
)

var (
	forceDownload bool
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "모델 파일 다운로드",
	Long:  `HuggingFace Hub에서 ONNX 모델 파일을 다운로드한다.`,
	RunE:  runDownload,
}

func init() {
	downloadCmd.Flags().BoolVarP(&forceDownload, "force", "f", false, "기존 파일 덮어쓰기")
	rootCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Embedder Worker - Model Download ===")
	fmt.Println()

	cfg := GetConfig()

	fmt.Println("[1] 설정 확인...")
	fmt.Printf("    Model repo: %s\n", cfg.ModelRepo)
	fmt.Printf("    Cache dir: %s\n", cfg.CacheDir)
	fmt.Println()

	// 강제 다운로드 시 기존 파일 삭제
	if forceDownload {
		fmt.Println("[1.5] 기존 파일 삭제 (--force)...")
		for _, filename := range downloader.ModelFiles {
			filePath := cfg.CacheDir + "/" + filename
			if err := os.Remove(filePath); err == nil {
				fmt.Printf("    삭제: %s\n", filename)
			}
		}
		fmt.Println()
	}

	fmt.Println("[2] 모델 파일 다운로드...")
	dl := downloader.NewHuggingFaceDownloader(cfg.HuggingFaceToken, cfg.ModelRepo, cfg.CacheDir)
	if err := dl.Download(); err != nil {
		return fmt.Errorf("다운로드 실패: %w", err)
	}
	fmt.Println()

	fmt.Println("[3] 다운로드된 파일 확인...")
	modelPath := dl.GetModelPath()
	tokenizerPath := dl.GetTokenizerPath()

	if info, err := os.Stat(modelPath); err == nil {
		fmt.Printf("    [OK] model.onnx: %.2f KB\n", float64(info.Size())/1024)
	}

	dataPath := modelPath + ".data"
	if info, err := os.Stat(dataPath); err == nil {
		fmt.Printf("    [OK] model.onnx.data: %.2f GB\n", float64(info.Size())/(1024*1024*1024))
	}

	if info, err := os.Stat(tokenizerPath); err == nil {
		fmt.Printf("    [OK] tokenizer.json: %.2f MB\n", float64(info.Size())/(1024*1024))
	}
	fmt.Println()

	fmt.Println("=== Download Completed ===")
	return nil
}
