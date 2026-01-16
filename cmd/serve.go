package cmd

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"

	"github.com/Whale0928/embedding-worker/internal/config"
	"github.com/Whale0928/embedding-worker/pkg/handler"
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

	// 2. Echo 서버 설정
	fmt.Println("[2] HTTP 서버 설정...")
	e := echo.New()
	e.HideBanner = true

	// 미들웨어
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 라우터 등록
	registerRoutes(e)
	fmt.Println("    [OK] 라우터 등록 완료")
	fmt.Println()

	// 3. 서버 시작
	addr := fmt.Sprintf(":%s", cfg.Qdrant.Port) // 일단 Qdrant 포트 재활용, 나중에 SERVER_PORT로 변경
	// TODO: SERVER_PORT 환경변수 추가
	addr = ":8080" // 임시 하드코딩

	fmt.Printf("[3] 서버 시작: http://localhost%s\n", addr)
	fmt.Println()

	return e.Start(addr)
}

func registerRoutes(e *echo.Echo) {
	// Health check
	healthHandler := handler.NewHealthHandler()
	healthHandler.Register(e)

	// API 그룹
	// api := e.Group("/api/v1")
	// TODO: 라우터 추가
}
