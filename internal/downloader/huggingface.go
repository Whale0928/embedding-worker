package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type HuggingFaceDownloader struct {
	token    string
	repo     string
	cacheDir string
}

func NewHuggingFaceDownloader(token, repo, cacheDir string) *HuggingFaceDownloader {
	return &HuggingFaceDownloader{
		token:    token,
		repo:     repo,
		cacheDir: cacheDir,
	}
}

// ModelFiles ONNX 모델에 필요한 파일 목록
var ModelFiles = []string{
	"model.onnx",
	"model.onnx.data",
	"tokenizer.json",
	"tokenizer_config.json",
	"special_tokens_map.json",
}

// Download 모든 모델 파일을 다운로드
func (d *HuggingFaceDownloader) Download() error {
	// 캐시 디렉토리 생성
	if err := os.MkdirAll(d.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	for _, filename := range ModelFiles {
		localPath := filepath.Join(d.cacheDir, filename)

		// 이미 존재하면 스킵
		if _, err := os.Stat(localPath); err == nil {
			fmt.Printf("[SKIP] %s already exists\n", filename)
			continue
		}

		fmt.Printf("[DOWNLOAD] %s ...\n", filename)
		if err := d.downloadFile(filename, localPath); err != nil {
			return fmt.Errorf("failed to download %s: %w", filename, err)
		}
		fmt.Printf("[OK] %s downloaded\n", filename)
	}

	return nil
}

// downloadFile 단일 파일 다운로드
func (d *HuggingFaceDownloader) downloadFile(filename, localPath string) error {
	url := fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", d.repo, filename)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Private 저장소 인증
	req.Header.Set("Authorization", "Bearer "+d.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// 임시 파일로 먼저 다운로드
	tmpPath := localPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	// 진행률 표시를 위한 Writer
	written, err := io.Copy(out, &progressReader{
		reader: resp.Body,
		total:  resp.ContentLength,
		name:   filename,
	})
	out.Close()

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	fmt.Printf("\n  -> %d bytes written\n", written)

	// 다운로드 완료 후 이름 변경
	return os.Rename(tmpPath, localPath)
}

// GetModelPath 모델 파일 경로 반환
func (d *HuggingFaceDownloader) GetModelPath() string {
	return filepath.Join(d.cacheDir, "model.onnx")
}

// GetTokenizerPath 토크나이저 파일 경로 반환
func (d *HuggingFaceDownloader) GetTokenizerPath() string {
	return filepath.Join(d.cacheDir, "tokenizer.json")
}

// progressReader 다운로드 진행률 표시
type progressReader struct {
	reader      io.Reader
	total       int64
	downloaded  int64
	name        string
	lastPercent int
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	if pr.total > 0 {
		percent := int(float64(pr.downloaded) / float64(pr.total) * 100)
		if percent != pr.lastPercent && percent%10 == 0 {
			fmt.Printf("  %s: %d%%\n", pr.name, percent)
			pr.lastPercent = percent
		}
	}

	return n, err
}
