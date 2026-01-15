package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Whale0928/embedding-worker/internal/config"
)

var (
	// 전역 설정
	cfg     *config.Config
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "embedder-worker",
	Short: "Go 기반 임베딩 워커",
	Long: `ONNX Runtime을 사용한 Go 기반 임베딩 워커.
KURE-v1 한국어 임베딩 모델을 사용하여 텍스트를 1024차원 벡터로 변환한다.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 모든 서브커맨드 실행 전에 설정 로드
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("설정 로드 실패: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "상세 로그 출력")
}

// GetConfig 설정 반환
func GetConfig() *config.Config {
	return cfg
}

// IsVerbose verbose 모드 여부
func IsVerbose() bool {
	return verbose
}
