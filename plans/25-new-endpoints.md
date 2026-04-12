# Spec 25: Image Generation, Audio & Rerank Endpoints

## Status
pending

## Goal
Add three new endpoint types: image generation (`POST /v1/images/generations`), audio transcription/TTS (`POST /v1/audio/transcriptions`, `POST /v1/audio/speech`), and reranking (`POST /v1/rerank`). Each introduces a new capability class beyond chat completions.

## Background
The gateway's current endpoint surface is: chat completions, embeddings, completions (spec 04). These new endpoints require new provider interfaces and handler types distinct from the chat path.

## Requirements

### 1. Image Generation

**Endpoint:** `POST /v1/images/generations`

**New types (`pkg/llm/types.go`):**
```go
type ImageGenerationRequest struct {
    Model          string `json:"model"`
    Prompt         string `json:"prompt"`
    N              int    `json:"n,omitempty"`           // default 1
    Size           string `json:"size,omitempty"`         // "256x256" | "512x512" | "1024x1024" | "1792x1024" | "1024x1792"
    Quality        string `json:"quality,omitempty"`      // "standard" | "hd"
    Style          string `json:"style,omitempty"`        // "vivid" | "natural"
    ResponseFormat string `json:"response_format,omitempty"` // "url" | "b64_json"
    User           string `json:"user,omitempty"`
}

type ImageData struct {
    URL          string `json:"url,omitempty"`
    B64JSON      string `json:"b64_json,omitempty"`
    RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type ImageGenerationResponse struct {
    Created int64       `json:"created"`
    Data    []ImageData `json:"data"`
}
```

**New provider interface (`pkg/llm/provider.go`):**
```go
type ImageProvider interface {
    Provider
    GenerateImage(ctx context.Context, req *ImageGenerationRequest) (*ImageGenerationResponse, error)
}
```

**OpenAI implementation** (`internal/proxy/provider_openai.go` extension):
- POST to `{baseURL}/images/generations`
- Pass request body as-is (already OpenAI format)
- Implement `GenerateImage` on `OpenAIProvider`

**Handler** (`internal/proxy/handler_images.go`):
- Parse request, resolve provider (same model resolution as chat)
- Check provider implements `ImageProvider`; return 501 if not
- Forward request, return response
- Audit log: prompt_tokens = 0, cost_usd = lookup from `image_pricing` (new table or static map by model+size+quality)

### 2. Audio Transcription

**Endpoint:** `POST /v1/audio/transcriptions`

Request is `multipart/form-data` with:
- `file` — audio file
- `model` — e.g., `"whisper-1"`
- `language` — optional ISO-639-1
- `prompt` — optional
- `response_format` — `"json"` | `"text"` | `"srt"` | `"vtt"`

**New types:**
```go
type AudioTranscriptionRequest struct {
    Model          string
    Language       string
    Prompt         string
    ResponseFormat string
    // File is passed as multipart — not a struct field
}

type AudioTranscriptionResponse struct {
    Text string `json:"text"`
}
```

**New provider interface:**
```go
type AudioProvider interface {
    Provider
    Transcribe(ctx context.Context, req *AudioTranscriptionRequest, audio io.Reader, filename string) (*AudioTranscriptionResponse, error)
}
```

**OpenAI implementation:** forward the multipart form to OpenAI's `/audio/transcriptions`.

**Handler** (`internal/proxy/handler_audio.go`):
- Parse multipart form (max 25MB — OpenAI's limit)
- Check provider implements `AudioProvider`
- Forward and return response

### 3. Text-to-Speech

**Endpoint:** `POST /v1/audio/speech`

**New types:**
```go
type TextToSpeechRequest struct {
    Model          string  `json:"model"`           // "tts-1" | "tts-1-hd"
    Input          string  `json:"input"`
    Voice          string  `json:"voice"`            // "alloy" | "echo" | "fable" | "onyx" | "nova" | "shimmer"
    ResponseFormat string  `json:"response_format,omitempty"` // "mp3" | "opus" | "aac" | "flac" | "wav" | "pcm"
    Speed          float64 `json:"speed,omitempty"` // 0.25 to 4.0
}
```

Response is binary audio stream (`Content-Type: audio/mpeg` etc.).

**Handler:** streams the binary audio response directly to the client.

### 4. Reranking

**Endpoint:** `POST /v1/rerank`

This is not an OpenAI endpoint — it follows the Cohere rerank API format, which has become a de-facto standard.

**New types:**
```go
type RerankRequest struct {
    Model     string   `json:"model"`
    Query     string   `json:"query"`
    Documents []string `json:"documents"`
    TopN      int      `json:"top_n,omitempty"`
}

type RerankResult struct {
    Index          int     `json:"index"`
    RelevanceScore float64 `json:"relevance_score"`
    Document       string  `json:"document,omitempty"`
}

type RerankResponse struct {
    ID      string        `json:"id"`
    Results []RerankResult `json:"results"`
    Usage   *Usage        `json:"usage,omitempty"`
}
```

**New provider interface:**
```go
type RerankProvider interface {
    Provider
    Rerank(ctx context.Context, req *RerankRequest) (*RerankResponse, error)
}
```

**Cohere implementation** (extend spec 13's `CohereProvider` to implement `RerankProvider`):
- POST to `https://api.cohere.com/v2/rerank`

### 5. Routing

All new endpoints use the same model-alias-based routing as chat completions. The handler checks if the resolved provider implements the required interface (e.g., `ImageProvider`) and returns HTTP 501 if not.

### 6. Audit logging

All new endpoint types are logged to `audit_log`. Extend `audit_log` with an `endpoint_type` column:
```sql
ALTER TABLE audit_log ADD COLUMN endpoint_type TEXT NOT NULL DEFAULT 'chat';
```

Values: `'chat'`, `'embeddings'`, `'completions'`, `'images'`, `'audio_transcription'`, `'tts'`, `'rerank'`.

## Out of Scope
- Image editing (`/v1/images/edits`) and variations
- Video generation
- Audio translation (`/v1/audio/translations`)
- Batch image generation
- Image generation for non-OpenAI providers (Stability AI) in this spec

## Acceptance Criteria
- [ ] `POST /v1/images/generations` with an OpenAI key and `dall-e-3` model returns image URLs
- [ ] `POST /v1/audio/transcriptions` with an audio file returns a transcription
- [ ] `POST /v1/audio/speech` returns audio binary
- [ ] `POST /v1/rerank` with a Cohere key returns ranked results
- [ ] 501 is returned when a model's provider doesn't support the endpoint type
- [ ] All requests are logged to `audit_log` with the correct `endpoint_type`

## Key Files
- `pkg/llm/types.go` — new request/response types
- `pkg/llm/provider.go` — `ImageProvider`, `AudioProvider`, `RerankProvider` interfaces
- `internal/proxy/handler_images.go` — new file
- `internal/proxy/handler_audio.go` — new file
- `internal/proxy/handler_rerank.go` — new file
- `internal/proxy/provider_openai.go` — implement `ImageProvider`, `AudioProvider`
- `internal/proxy/provider_cohere.go` — implement `RerankProvider` (spec 13 extension)
- `internal/api/router.go` — new routes
- `internal/storage/migrations/` — `endpoint_type` column migration
