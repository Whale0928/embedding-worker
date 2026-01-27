package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Whale0928/embedding-worker/pkg/repository"
)

// VectorHandler 벡터 관련 HTTP 핸들러
type VectorHandler struct {
	vespa *repository.VespaClient
}

// NewVectorHandler 생성자
func NewVectorHandler(vespa *repository.VespaClient) *VectorHandler {
	return &VectorHandler{
		vespa: vespa,
	}
}

// Register 라우터 등록
func (h *VectorHandler) Register(e *echo.Echo) {
	e.GET("/vector", h.ListKeys)
}

// ListKeys 저장된 문서 key 목록 조회
func (h *VectorHandler) ListKeys(c echo.Context) error {
	schema := repository.Schema{Namespace: "sample", DocType: "sample_vector"}
	result, err := h.vespa.ListDocuments(schema, 100)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	keys := make([]string, 0)
	for _, doc := range result.Documents {
		keys = append(keys, doc.ID)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"keys":  keys,
		"count": len(keys),
	})
}
