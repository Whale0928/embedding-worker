package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	HuggingFaceToken string
	ModelRepo        string
	CacheDir         string
}

func Load() (*Config, error) {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		// .env 파일이 없어도 환경변수에서 읽을 수 있음
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	token := os.Getenv("HUGGING_FACE_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("HUGGING_FACE_TOKEN is required")
	}

	modelRepo := os.Getenv("HUGGING_FACE_MODEL_REPO")
	if modelRepo == "" {
		return nil, fmt.Errorf("HUGGING_FACE_MODEL_REPO is required")
	}

	// 캐시 디렉토리 설정
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, ".cache", "embedding-worker")

	return &Config{
		HuggingFaceToken: token,
		ModelRepo:        modelRepo,
		CacheDir:         cacheDir,
	}, nil
}
