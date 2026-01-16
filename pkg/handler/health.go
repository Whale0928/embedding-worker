package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler 헬스체크 핸들러
type HealthHandler struct{}

// NewHealthHandler 생성자
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Register 라우터 등록
func (h *HealthHandler) Register(e *echo.Echo) {
	e.GET("/health", h.Health)
}

// Health 헬스체크 응답
func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
