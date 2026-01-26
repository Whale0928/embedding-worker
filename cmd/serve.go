package cmd

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"

	"github.com/Whale0928/embedding-worker/internal/config"
	"github.com/Whale0928/embedding-worker/pkg/handler"
	"github.com/Whale0928/embedding-worker/pkg/repository"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "HTTP 서버 시작",
	Long:  `임베딩 워커 HTTP 서버를 시작한다.`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Embedder Worker - Server ===")
	fmt.Println()

	cfg := GetConfig()

	// 1. DB 연결
	fmt.Println("[1] DB 연결 중...")
	db, err := config.NewDB(&cfg.DB)
	if err != nil {
		return fmt.Errorf("DB 연결 실패: %w", err)
	}

	// DB 연결 확인
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("DB 인스턴스 조회 실패: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("DB Ping 실패: %w", err)
	}
	fmt.Printf("    [OK] DB 연결 성공: %s:%s/%s\n", cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	fmt.Println()

	// 2. Vespa 클라이언트 생성
	fmt.Println("[2] Vespa 클라이언트 설정...")
	vespaClient := repository.NewVespaClient(
		fmt.Sprintf("http://%s:%s", cfg.Vector.Host, cfg.Vector.Port),
		"sample",        // <- cfg 조절로 변경
		"sample_vector", // <- docType 채우기
	)
	fmt.Println("    [OK] Vespa 클라이언트 생성 완료")
	fmt.Println()

	// 3. Echo 서버 설정
	fmt.Println("[3] HTTP 서버 설정...")
	e := echo.New()
	e.HideBanner = true
	defer func() { _ = e.Close() }()

	// 미들웨어
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 라우터 등록
	registerRoutes(e, vespaClient)
	fmt.Println("    [OK] 라우터 등록 완료")
	fmt.Println()

	// 4. 서버 시작
	addr := fmt.Sprintf(":%s", cfg.HttpConfig.Port)
	fmt.Printf("[4] 서버 시작: http://localhost%s\n", addr)
	fmt.Println()

	return e.Start(addr)
}

func registerRoutes(e *echo.Echo, vespaClient *repository.VespaClient) {
	// Health check
	healthHandler := handler.NewHealthHandler()
	vectorHandler := handler.NewVectorHandler(vespaClient)

	healthHandler.Register(e)
	vectorHandler.Register(e)
}
