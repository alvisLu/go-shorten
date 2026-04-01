package whisper

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"sync"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Whisper wraps a loaded whisper.cpp model.
type Whisper struct {
	mu    sync.Mutex
	model whisper.Model
}

// LoadModel 從 ggml 模型檔路徑載入模型（CGO 阻塞）。
func LoadModel(modelPath string) (*Whisper, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("whisper: load model %q: %w", modelPath, err)
	}
	log.Printf("whisper: loaded model from %q\n", modelPath)
	return &Whisper{model: model}, nil
}

// Close 釋放模型資源。
func (w *Whisper) Close() {
	w.model.Close()
}

// Transcribe 對 16kHz Float32 mono PCM 執行 STT，回傳辨識文字。
// lang 為 BCP-47 短代碼（如 "zh"、"en"）；傳空字串則自動偵測。
func (w *Whisper) Transcribe(pcm []float32, lang string) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	ctx, err := w.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("whisper: new context: %w", err)
	}

	if lang == "" {
		lang = "auto"
	}
	if err := ctx.SetLanguage(lang); err != nil {
		_ = ctx.SetLanguage("auto")
	}

	// 與 Node.js STT_BASE_CONFIG 對齊（bindings 僅暴露部分參數）
	ctx.SetBeamSize(5)
	ctx.SetTemperature(0)
	ctx.SetEntropyThold(2.4) // compression_ratio_threshold
	ctx.SetMaxContext(0)

	if err := ctx.Process(pcm, nil, nil, nil); err != nil {
		return "", fmt.Errorf("whisper: process: %w", err)
	}

	var sb strings.Builder
	for {
		seg, err := ctx.NextSegment()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return "", fmt.Errorf("whisper: next segment: %w", err)
			}
			break
		}
		sb.WriteString(seg.Text)
	}

	text := strings.TrimSpace(sb.String())

	log.Printf("transcribed text: %q\n", text)
	return text, nil
}

// Resample 將 mono Float32 PCM 從 fromRate 線性插值降採樣至 16000 Hz。
// 若 fromRate 已為 16000，直接回傳原 slice（不複製）。
func Resample(input []float32, fromRate int) []float32 {
	if fromRate == 16000 {
		return input
	}
	ratio := float64(fromRate) / 16000.0
	length := int(math.Ceil(float64(len(input)) / ratio))
	output := make([]float32, length)
	for i := range output {
		pos := float64(i) * ratio
		idx := int(pos)
		frac := float32(pos - float64(idx))
		if idx+1 < len(input) {
			output[i] = input[idx]*(1-frac) + input[idx+1]*frac
		} else {
			output[i] = input[idx]
		}
	}
	return output
}
