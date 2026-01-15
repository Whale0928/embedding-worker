package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	ort "github.com/yalue/onnxruntime_go"

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

	// 4. ONNX Runtime으로 모델 로드 및 검증
	fmt.Println("[4] Loading model with ONNX Runtime...")
	if err := validateONNXModel(modelPath); err != nil {
		log.Fatalf("Failed to validate ONNX model: %v", err)
	}

	fmt.Println()
	fmt.Println("=== All Validations Passed ===")
}

// =============================================================================
// ONNX 모델 검증 함수 (나중에 제거 가능)
// =============================================================================

func validateONNXModel(modelPath string) error {
	fmt.Println()
	fmt.Println("    ┌─────────────────────────────────────────────────┐")
	fmt.Println("    │          ONNX Runtime Model Validation          │")
	fmt.Println("    └─────────────────────────────────────────────────┘")
	fmt.Println()

	// Step 1: ONNX Runtime 라이브러리 경로 설정
	fmt.Println("    [Step 1] Setting ONNX Runtime library path...")
	libPath := getONNXRuntimeLibPath()
	fmt.Printf("             OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("             Library path: %s\n", libPath)

	if _, err := os.Stat(libPath); err != nil {
		return fmt.Errorf("ONNX Runtime library not found at %s", libPath)
	}
	fmt.Println("             [OK] Library file exists")

	ort.SetSharedLibraryPath(libPath)
	fmt.Println("             [OK] Library path set")
	fmt.Println()

	// Step 2: ONNX Runtime 환경 초기화
	fmt.Println("    [Step 2] Initializing ONNX Runtime environment...")
	err := ort.InitializeEnvironment()
	if err != nil {
		return fmt.Errorf("failed to initialize ONNX Runtime: %w", err)
	}
	defer func() {
		fmt.Println()
		fmt.Println("    [Cleanup] Destroying ONNX Runtime environment...")
		err := ort.DestroyEnvironment()
		if err != nil {
			fmt.Printf("             [WARN] Failed to destroy environment: %v\n", err)
		} else {
			fmt.Println("             [OK] Environment destroyed")
		}
	}()
	fmt.Println("             [OK] Environment initialized")
	fmt.Println()

	// Step 3: 모델 파일 정보 확인
	fmt.Println("    [Step 3] Checking model file...")
	fmt.Printf("             Model path: %s\n", modelPath)

	modelInfo, err := os.Stat(modelPath)
	if err != nil {
		return fmt.Errorf("cannot stat model file: %w", err)
	}
	fmt.Printf("             Model size: %d bytes (%.2f KB)\n", modelInfo.Size(), float64(modelInfo.Size())/1024)

	// 외부 데이터 파일 확인
	dataPath := modelPath + ".data"
	if dataInfo, err := os.Stat(dataPath); err == nil {
		fmt.Printf("             External data: %s (%.2f GB)\n", dataPath, float64(dataInfo.Size())/(1024*1024*1024))
	}
	fmt.Println("             [OK] Model files verified")
	fmt.Println()

	// Step 4: 입출력 Shape 정의
	fmt.Println("    [Step 4] Defining input/output shapes...")

	// KURE-v1 모델 입력: [batch_size, sequence_length]
	// 테스트용으로 batch=1, seq_len=8 사용
	batchSize := int64(1)
	seqLen := int64(8)

	inputShape := ort.NewShape(batchSize, seqLen)
	fmt.Printf("             Input shape: [%d, %d] (batch_size, sequence_length)\n", batchSize, seqLen)

	// 출력 shape: last_hidden_state [batch, seq, 1024], pooler_output [batch, 1024]
	outputShape1 := ort.NewShape(batchSize, seqLen, 1024)
	outputShape2 := ort.NewShape(batchSize, 1024)
	fmt.Printf("             Output1 shape (last_hidden_state): [%d, %d, 1024]\n", batchSize, seqLen)
	fmt.Printf("             Output2 shape (pooler_output): [%d, 1024]\n", batchSize)
	fmt.Println("             [OK] Shapes defined")
	fmt.Println()

	// Step 5: 입력 텐서 생성
	fmt.Println("    [Step 5] Creating input tensors...")

	// 더미 input_ids (토큰 ID들)
	inputIDs := make([]int64, batchSize*seqLen)
	for i := range inputIDs {
		inputIDs[i] = int64(i + 1) // 더미 토큰 ID: 1, 2, 3, ...
	}
	fmt.Printf("             input_ids: %v\n", inputIDs)

	inputIDsTensor, err := ort.NewTensor(inputShape, inputIDs)
	if err != nil {
		return fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Destroy()
	fmt.Println("             [OK] input_ids tensor created")

	// attention_mask (모든 토큰 활성화)
	attentionMask := make([]int64, batchSize*seqLen)
	for i := range attentionMask {
		attentionMask[i] = 1 // 모든 토큰 활성화
	}
	fmt.Printf("             attention_mask: %v\n", attentionMask)

	attentionMaskTensor, err := ort.NewTensor(inputShape, attentionMask)
	if err != nil {
		return fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Destroy()
	fmt.Println("             [OK] attention_mask tensor created")
	fmt.Println()

	// Step 6: 출력 텐서 생성
	fmt.Println("    [Step 6] Creating output tensors...")

	// last_hidden_state 출력
	output1Data := make([]float32, batchSize*seqLen*1024)
	output1Tensor, err := ort.NewTensor(outputShape1, output1Data)
	if err != nil {
		return fmt.Errorf("failed to create output1 tensor: %w", err)
	}
	defer output1Tensor.Destroy()
	fmt.Println("             [OK] last_hidden_state tensor created")

	// pooler_output 출력
	output2Data := make([]float32, batchSize*1024)
	output2Tensor, err := ort.NewTensor(outputShape2, output2Data)
	if err != nil {
		return fmt.Errorf("failed to create output2 tensor: %w", err)
	}
	defer output2Tensor.Destroy()
	fmt.Println("             [OK] pooler_output tensor created")
	fmt.Println()

	// Step 7: 세션 생성
	fmt.Println("    [Step 7] Creating ONNX session...")
	fmt.Println("             This may take a while for large models...")

	session, err := ort.NewAdvancedSession(
		modelPath,
		[]string{"input_ids", "attention_mask"},
		[]string{"last_hidden_state", "pooler_output"},
		[]ort.ArbitraryTensor{inputIDsTensor, attentionMaskTensor},
		[]ort.ArbitraryTensor{output1Tensor, output2Tensor},
		nil, // 기본 옵션 사용
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Destroy()
	fmt.Println("             [OK] Session created successfully!")
	fmt.Println()

	// Step 8: 추론 실행
	fmt.Println("    [Step 8] Running inference...")
	err = session.Run()
	if err != nil {
		return fmt.Errorf("failed to run inference: %w", err)
	}
	fmt.Println("             [OK] Inference completed!")
	fmt.Println()

	// Step 9: 결과 확인
	fmt.Println("    [Step 9] Checking results...")

	// pooler_output 결과 확인 (임베딩 벡터)
	embedding := output2Tensor.GetData()
	fmt.Printf("             Embedding dimension: %d\n", len(embedding))
	fmt.Printf("             First 5 values: [%.6f, %.6f, %.6f, %.6f, %.6f]\n",
		embedding[0], embedding[1], embedding[2], embedding[3], embedding[4])
	fmt.Printf("             Last 5 values: [%.6f, %.6f, %.6f, %.6f, %.6f]\n",
		embedding[1019], embedding[1020], embedding[1021], embedding[1022], embedding[1023])

	// 결과 유효성 검증 (모두 0이 아닌지 확인)
	nonZeroCount := 0
	for _, v := range embedding {
		if v != 0 {
			nonZeroCount++
		}
	}
	fmt.Printf("             Non-zero values: %d / %d\n", nonZeroCount, len(embedding))

	if nonZeroCount == 0 {
		return fmt.Errorf("all embedding values are zero - model may not be working correctly")
	}
	fmt.Println("             [OK] Embedding values look valid!")
	fmt.Println()

	fmt.Println("    ┌─────────────────────────────────────────────────┐")
	fmt.Println("    │         ONNX Model Validation PASSED!          │")
	fmt.Println("    └─────────────────────────────────────────────────┘")

	return nil
}

// getONNXRuntimeLibPath OS별 ONNX Runtime 라이브러리 경로 반환
func getONNXRuntimeLibPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS (Homebrew)
		// Apple Silicon
		if runtime.GOARCH == "arm64" {
			return "/opt/homebrew/opt/onnxruntime/lib/libonnxruntime.dylib"
		}
		// Intel Mac
		return "/usr/local/opt/onnxruntime/lib/libonnxruntime.dylib"
	case "linux":
		return "/usr/lib/libonnxruntime.so"
	case "windows":
		return "onnxruntime.dll"
	default:
		return "libonnxruntime.so"
	}
}
