# Real-time Speech Translation

即時語音轉錄與翻譯服務後端。

透過 WebSocket 接收雙聲道音訊（麥克風 + 系統音訊），以 Whisper.cpp 進行語音辨識，再經由翻譯模組(openAI/google translate) API 輸出翻譯結果。

## Tech Stack

| Category                | Technology                                                                            |
| ----------------------- | ------------------------------------------------------------------------------------- |
| Language                | Go (Gin framework)                                                                    |
| Real-time Communication | WebSocket (gorilla/websocket)                                                         |
| Speech-to-Text          | Whisper.cpp — 透過 CGO 靜態連結，直接在 Go 中呼叫 C++ 推論                            |
| Translation API         | OpenAI API / Google Translate                                                         |
| Containerization        | Docker multi-stage build — 編譯 Whisper.cpp 靜態庫 + Go binary，產出單一 Alpine image |
| Orchestration           | Docker Compose — 含資源限制 (CPU/Memory) 設定                                         |
| Database                | PostgreSQL                                                                            |
| Audio Processing        | PCM Float32 LE 接收、線性內插重取樣至 16kHz、RMS 靜音過濾                             |

## Architecture

```
Client (WebSocket)
  │  Binary: [isFinal][channel][id][PCM Float32 LE]
  │  Text:   {"type":"start","sourceLang":"zh","targetLang":"en","sampleRate":48000}
  ▼
Gateway (Gin + WebSocket /ws)
  ▼
Session (per-connection state, per-channel buffer)
  ▼
STT Pipeline
  ├─ Interim (200ms debounce) → partial transcript + translation
  └─ Final (segment end)      → complete transcript + translation
  ▼
Whisper.cpp (local, 16kHz resample)
  ▼
Translator (OpenAI Responses API / DSE/GX — HTTP)
  ▼
Client ← JSON: {"type":"transcript"|"translation", "channel":"mic"|"loopback", "text":"...", "final":true}
```

**Channels**: `mic` (channel 0), `loopback` (channel 1)

## WebSocket Protocol

### Control Messages (Text JSON)

```jsonc
// 開始辨識
{"type": "start", "sourceLang": "zh", "targetLang": "en", "sampleRate": 48000, "enableDenoise": true}

// 停止辨識
{"type": "stop"}

// 健康檢查
{"type": "health"}
```

### Audio Frame (Binary)

| Offset | Size     | Description                        |
| ------ | -------- | ---------------------------------- |
| 0      | 1 byte   | `isFinal` (0 = interim, 1 = final) |
| 1      | 1 byte   | channel (0 = mic, 1 = loopback)    |
| 2      | 21 bytes | segment ID (zero-padded)           |
| 23     | variable | PCM Float32 LE samples             |

### Response Messages (Text JSON)

```jsonc
{"type": "transcript",   "channel": "mic", "id": "seg001", "text": "你好", "final": false}
{"type": "translation",  "channel": "mic", "id": "seg001", "text": "hello", "final": true}
{"status": "started"}
{"error": "something went wrong"}
```

## Environment Variables

| Variable             | Description                           | Default                 |
| -------------------- | ------------------------------------- | ----------------------- |
| `HOST`               | Server bind address                   | `localhost`             |
| `PORT`               | Server port                           | `8080`                  |
| `DATABASE_URL`       | PostgreSQL connection string          | -                       |
| `ALLOWED_ORIGINS`    | CORS allowed origins                  | `http://localhost:3000` |
| `WHISPER_MODEL_PATH` | Whisper model path (未設定則停用 STT) | -                       |
| `DSE_GX_URL`         | DSE/GX translation API endpoint       | -                       |
| `OPENAI_API_KEY`     | OpenAI API key                        | -                       |
| `OPENAI_MODEL`       | OpenAI model name                     | `gpt-4o`                |
| `OPENAI_URL`         | OpenAI Responses API endpoint         | -                       |

## OpenAI Translation

使用 OpenAI **Responses API** 進行翻譯。透過 `Translator` interface 抽象翻譯層，可在 OpenAI 與 DSE/GX 之間切換。

**Request** — 以 system prompt 指定來源/目標語言，將辨識文字作為 user input 送出：

```json
POST ${OPENAI_URL}
Authorization: Bearer ${OPENAI_API_KEY}

{
  "model": "gpt-4o",
  "instructions": "You are a professional translator. Translate the following text from zh to en. Output only the translated text, no explanations.",
  "input": [
    {"role": "user", "content": "你好"}
  ]
}
```

**Response** — 從 `output[].content[].text` 取出翻譯結果：

```json
{
  "output": [
    {
      "type": "message",
      "content": [{"type": "output_text", "text": "Hello"}]
    }
  ]
}
```

## Run

```bash
# Local
go run ./cmd/server

# Docker
docker compose up --build
```

Docker 首次啟動時，若 model 不存在會自動從 HuggingFace 下載。

> Docker 環境須設定 `HOST=0.0.0.0`。

## Whisper Models

| Model            | Size    |
| ---------------- | ------- |
| `ggml-tiny.bin`  | ~75 MB  |
| `ggml-base.bin`  | ~148 MB |
| `ggml-small.bin` | ~488 MB |

Models 來源: [huggingface.co/ggerganov/whisper.cpp](https://huggingface.co/ggerganov/whisper.cpp)

## Demo

https://github.com/user-attachments/assets/743cc780-6485-4ef4-a6b9-f3cf42df7155

Test video: [How My Kids Learn English](https://www.youtube.com/watch?v=pfQoC8n4XxY&t=53s)
