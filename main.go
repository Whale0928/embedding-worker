package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Whale0928/embedding-worker/internal/config"
	"github.com/Whale0928/embedding-worker/internal/downloader"
)

func main() {
	fmt.Println("=== Embedding Worker ===")
	fmt.Println()

	// 1. 설정 로드
	fmt.Println("[1] Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("    Model repo: %s\n", cfg.ModelRepo)
	fmt.Printf("    Cache dir: %s\n", cfg.CacheDir)
	fmt.Println()

	// 2. 모델 다운로드
	fmt.Println("[2] Downloading model files...")
	dl := downloader.NewHuggingFaceDownloader(cfg.HuggingFaceToken, cfg.ModelRepo, cfg.CacheDir)
	if err := dl.Download(); err != nil {
		log.Fatalf("Failed to download model: %v", err)
	}
	fmt.Println()

	// 3. 파일 검증
	fmt.Println("[3] Validating downloaded files...")
	modelPath := dl.GetModelPath()
	tokenizerPath := dl.GetTokenizerPath()

	if _, err := os.Stat(modelPath); err != nil {
		log.Fatalf("Model file not found: %s", modelPath)
	}
	fmt.Printf("    [OK] model.onnx exists: %s\n", modelPath)

	if _, err := os.Stat(tokenizerPath); err != nil {
		log.Fatalf("Tokenizer file not found: %s", tokenizerPath)
	}
	fmt.Printf("    [OK] tokenizer.json exists: %s\n", tokenizerPath)

	// model.onnx.data 확인
	dataPath := modelPath + ".data"
	if info, err := os.Stat(dataPath); err == nil {
		fmt.Printf("    [OK] model.onnx.data exists: %.2f GB\n", float64(info.Size())/(1024*1024*1024))
	}

	fmt.Println()
	fmt.Println("=== Download Complete ===")
	fmt.Println()
	fmt.Println("Next step: Load model with ONNX Runtime")
	fmt.Println("Required: Install ONNX Runtime shared library")
	fmt.Println("  macOS: brew install onnxruntime")
	fmt.Println("  Linux: Download from https://github.com/microsoft/onnxruntime/releases")
}
