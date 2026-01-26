package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VespaClient Vespa Document API 클라이언트
type VespaClient struct {
	baseURL    string
	httpClient *http.Client
	namespace  string
	docType    string
}

// VespaDocument Visit API 응답의 개별 문서
type VespaDocument struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// VisitResponse Visit API 전체 응답
type VisitResponse struct {
	Documents     []VespaDocument `json:"documents"`
	DocumentCount int             `json:"documentCount"`
	Continuation  string          `json:"continuation,omitempty"`
}

func NewVespaClient(baseURL, namespace, docType string) *VespaClient {
	return &VespaClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		namespace:  namespace,
		docType:    docType,
	}
}

// ListDocuments 저장된 문서 목록 조회 (Visit API)
func (client *VespaClient) ListDocuments(count int) (*VisitResponse, error) {
	url := fmt.Sprintf("%s/document/v1/%s/%s/docid?wantedDocumentCount=%d", client.baseURL, client.namespace, client.docType, count)
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("vespa client  요청 실패: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vespa client response HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result VisitResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("vespa client response JSON 디코딩 실패: %w", err)
	}
	return &result, nil
}

// GetDocument 단일 문서 조회
// API: GET /document/v1/{namespace}/{docType}/docid/{id}
func (client *VespaClient) GetDocument(id string) (*VespaDocument, error) {
	// TODO: ListDocuments 참고해서 구현
	return nil, fmt.Errorf("not implemented")
}
