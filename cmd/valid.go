package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	ort "github.com/yalue/onnxruntime_go"

	"github.com/Whale0928/embedding-worker/internal/downloader"
)

var validCmd = &cobra.Command{
	Use:   "valid",
	Short: "ONNX 모델 검증",
	Long:  `다운로드된 ONNX 모델이 정상적으로 동작하는지 검증한다.`,
	RunE:  runValid,
}

func init() {
	rootCmd.AddCommand(validCmd)
}

func runValid(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Embedder Worker - Model Validation ===")
	fmt.Println()

	cfg := GetConfig()

	// 1. 다운로더 초기화 및 경로 확인
	fmt.Println("[1] 모델 경로 확인...")
	dl := downloader.NewHuggingFaceDownloader(cfg.HuggingFaceToken, cfg.ModelRepo, cfg.CacheDir)
	modelPath := dl.GetModelPath()
	tokenizerPath := dl.GetTokenizerPath()

	if _, err := os.Stat(modelPath); err != nil {
		return fmt.Errorf("모델 파일 없음: %s (먼저 'download' 명령 실행)", modelPath)
	}
	fmt.Printf("    model.onnx: %s\n", modelPath)

	if _, err := os.Stat(tokenizerPath); err != nil {
		return fmt.Errorf("토크나이저 파일 없음: %s", tokenizerPath)
	}
	fmt.Printf("    tokenizer.json: %s\n", tokenizerPath)

	// model.onnx.data 확인
	dataPath := modelPath + ".data"
	if info, err := os.Stat(dataPath); err == nil {
		fmt.Printf("    model.onnx.data: %.2f GB\n", float64(info.Size())/(1024*1024*1024))
	}
	fmt.Println()

	// 2. ONNX Runtime 모델 검증
	fmt.Println("[2] ONNX Runtime 모델 검증...")
	if err := validateONNXModel(modelPath); err != nil {
		return fmt.Errorf("모델 검증 실패: %w", err)
	}

	fmt.Println()
	fmt.Println("=== Validation Completed ===")
	return nil
}

func validateONNXModel(modelPath string) error {
	fmt.Println()
	fmt.Println("    +-----------------------------------------------------+")
	fmt.Println("    |          ONNX Runtime Model Validation              |")
	fmt.Println("    +-----------------------------------------------------+")
	fmt.Println()

	// Step 1: ONNX Runtime 라이브러리 경로 설정
	fmt.Println("    [Step 1] ONNX Runtime 라이브러리 경로 설정...")
	libPath := getONNXRuntimeLibPath()
	fmt.Printf("             OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("             Library path: %s\n", libPath)

	if _, err := os.Stat(libPath); err != nil {
		return fmt.Errorf("ONNX Runtime 라이브러리 없음: %s\n설치 방법: %s", libPath, getInstallHint())
	}
	fmt.Println("             [OK] 라이브러리 파일 존재")

	ort.SetSharedLibraryPath(libPath)
	fmt.Println("             [OK] 라이브러리 경로 설정 완료")
	fmt.Println()

	// Step 2: ONNX Runtime 환경 초기화
	fmt.Println("    [Step 2] ONNX Runtime 환경 초기화...")
	err := ort.InitializeEnvironment()
	if err != nil {
		return fmt.Errorf("ONNX Runtime 초기화 실패: %w", err)
	}
	defer func() {
		fmt.Println()
		fmt.Println("    [Cleanup] ONNX Runtime 환경 정리...")
		if err := ort.DestroyEnvironment(); err != nil {
			fmt.Printf("             [WARN] 환경 정리 실패: %v\n", err)
		} else {
			fmt.Println("             [OK] 환경 정리 완료")
		}
	}()
	fmt.Println("             [OK] 환경 초기화 완료")
	fmt.Println()

	// Step 3: 모델 파일 정보 확인
	fmt.Println("    [Step 3] 모델 파일 정보 확인...")
	fmt.Printf("             Model path: %s\n", modelPath)

	modelInfo, err := os.Stat(modelPath)
	if err != nil {
		return fmt.Errorf("모델 파일 확인 실패: %w", err)
	}
	fmt.Printf("             Model size: %d bytes (%.2f KB)\n", modelInfo.Size(), float64(modelInfo.Size())/1024)

	dataPath := modelPath + ".data"
	if dataInfo, err := os.Stat(dataPath); err == nil {
		fmt.Printf("             External data: %.2f GB\n", float64(dataInfo.Size())/(1024*1024*1024))
	}
	fmt.Println("             [OK] 모델 파일 확인 완료")
	fmt.Println()

	// Step 4: 입출력 Shape 정의
	fmt.Println("    [Step 4] 입출력 Shape 정의...")

	batchSize := int64(1)
	seqLen := int64(8)

	inputShape := ort.NewShape(batchSize, seqLen)
	fmt.Printf("             Input shape: [%d, %d] (batch_size, sequence_length)\n", batchSize, seqLen)

	outputShape1 := ort.NewShape(batchSize, seqLen, 1024)
	outputShape2 := ort.NewShape(batchSize, 1024)
	fmt.Printf("             Output1 shape (last_hidden_state): [%d, %d, 1024]\n", batchSize, seqLen)
	fmt.Printf("             Output2 shape (pooler_output): [%d, 1024]\n", batchSize)
	fmt.Println("             [OK] Shape 정의 완료")
	fmt.Println()

	// Step 5: 입력 텐서 생성
	fmt.Println("    [Step 5] 입력 텐서 생성...")

	inputIDs := make([]int64, batchSize*seqLen)
	for i := range inputIDs {
		inputIDs[i] = int64(i + 1)
	}
	if IsVerbose() {
		fmt.Printf("             input_ids: %v\n", inputIDs)
	}

	inputIDsTensor, err := ort.NewTensor(inputShape, inputIDs)
	if err != nil {
		return fmt.Errorf("input_ids 텐서 생성 실패: %w", err)
	}
	defer inputIDsTensor.Destroy()
	fmt.Println("             [OK] input_ids 텐서 생성 완료")

	attentionMask := make([]int64, batchSize*seqLen)
	for i := range attentionMask {
		attentionMask[i] = 1
	}
	if IsVerbose() {
		fmt.Printf("             attention_mask: %v\n", attentionMask)
	}

	attentionMaskTensor, err := ort.NewTensor(inputShape, attentionMask)
	if err != nil {
		return fmt.Errorf("attention_mask 텐서 생성 실패: %w", err)
	}
	defer attentionMaskTensor.Destroy()
	fmt.Println("             [OK] attention_mask 텐서 생성 완료")
	fmt.Println()

	// Step 6: 출력 텐서 생성
	fmt.Println("    [Step 6] 출력 텐서 생성...")

	output1Data := make([]float32, batchSize*seqLen*1024)
	output1Tensor, err := ort.NewTensor(outputShape1, output1Data)
	if err != nil {
		return fmt.Errorf("output1 텐서 생성 실패: %w", err)
	}
	defer output1Tensor.Destroy()
	fmt.Println("             [OK] last_hidden_state 텐서 생성 완료")

	output2Data := make([]float32, batchSize*1024)
	output2Tensor, err := ort.NewTensor(outputShape2, output2Data)
	if err != nil {
		return fmt.Errorf("output2 텐서 생성 실패: %w", err)
	}
	defer output2Tensor.Destroy()
	fmt.Println("             [OK] pooler_output 텐서 생성 완료")
	fmt.Println()

	// Step 7: 세션 생성
	fmt.Println("    [Step 7] ONNX 세션 생성...")
	fmt.Println("             대용량 모델 로딩 중...")

	session, err := ort.NewAdvancedSession(
		modelPath,
		[]string{"input_ids", "attention_mask"},
		[]string{"last_hidden_state", "pooler_output"},
		[]ort.ArbitraryTensor{inputIDsTensor, attentionMaskTensor},
		[]ort.ArbitraryTensor{output1Tensor, output2Tensor},
		nil,
	)
	if err != nil {
		return fmt.Errorf("세션 생성 실패: %w", err)
	}
	defer session.Destroy()
	fmt.Println("             [OK] 세션 생성 완료")
	fmt.Println()

	// Step 8: 추론 실행
	fmt.Println("    [Step 8] 추론 실행...")
	err = session.Run()
	if err != nil {
		return fmt.Errorf("추론 실패: %w", err)
	}
	fmt.Println("             [OK] 추론 완료")
	fmt.Println()

	// Step 9: 결과 확인
	fmt.Println("    [Step 9] 결과 확인...")

	embedding := output2Tensor.GetData()
	fmt.Printf("             Embedding dimension: %d\n", len(embedding))
	fmt.Printf("             First 5 values: [%.6f, %.6f, %.6f, %.6f, %.6f]\n",
		embedding[0], embedding[1], embedding[2], embedding[3], embedding[4])
	fmt.Printf("             Last 5 values: [%.6f, %.6f, %.6f, %.6f, %.6f]\n",
		embedding[1019], embedding[1020], embedding[1021], embedding[1022], embedding[1023])

	nonZeroCount := 0
	for _, v := range embedding {
		if v != 0 {
			nonZeroCount++
		}
	}
	fmt.Printf("             Non-zero values: %d / %d\n", nonZeroCount, len(embedding))

	if nonZeroCount == 0 {
		return fmt.Errorf("모든 임베딩 값이 0 - 모델 문제 가능성")
	}
	fmt.Println("             [OK] 임베딩 값 정상")
	fmt.Println()

	fmt.Println("    +-----------------------------------------------------+")
	fmt.Println("    |           ONNX Model Validation PASSED!             |")
	fmt.Println("    +-----------------------------------------------------+")

	return nil
}

func getONNXRuntimeLibPath() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "/opt/homebrew/opt/onnxruntime/lib/libonnxruntime.dylib"
		}
		return "/usr/local/opt/onnxruntime/lib/libonnxruntime.dylib"
	case "linux":
		return "/usr/lib/libonnxruntime.so"
	case "windows":
		return "onnxruntime.dll"
	default:
		return "libonnxruntime.so"
	}
}

func getInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install onnxruntime"
	case "linux":
		return "apt install libonnxruntime-dev 또는 https://github.com/microsoft/onnxruntime/releases"
	case "windows":
		return "https://github.com/microsoft/onnxruntime/releases 에서 다운로드"
	default:
		return "https://github.com/microsoft/onnxruntime/releases"
	}
}
