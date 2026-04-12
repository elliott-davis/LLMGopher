# Spec 13: Cohere Provider

## Status
pending

## Goal
Add a Cohere provider supporting chat completions, embeddings, and reranking. Cohere's Command R+ is a strong RAG-optimized model and its `embed-english-v3.0` embeddings are widely used; its rerank endpoint has no equivalent in other providers.

## Background
Cohere has its own API format (not OpenAI-compatible), but offers clear documentation. The provider implements `llm.Provider` for chat and optionally `llm.EmbeddingProvider` for embeddings. The rerank endpoint will be handled as part of spec 25.

Cohere API base URL: `https://api.cohere.com/v2`

Relevant endpoints:
- `POST /v2/chat` — chat completions
- `POST /v2/embed` — embeddings
- `POST /v2/rerank` — rerank (out of scope for this spec)

## Requirements

### 1. New provider (`internal/proxy/provider_cohere.go`)

```go
type CohereProvider struct {
    apiKey string
    client *http.Client
}

func NewCohereProvider(apiKey string) *CohereProvider
func (p *CohereProvider) Name() string // "cohere"
func (p *CohereProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error)
func (p *CohereProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error)
func (p *CohereProvider) EmbedContent(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error)
```

### 2. Chat request translation

`llm.ChatCompletionRequest` → Cohere v2 chat request:
```json
{
  "model": "command-r-plus",
  "messages": [
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": "Hi!"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "stream": false
}
```

- `role: "system"` → Cohere accepts system messages directly in the messages array (v2 API)
- `tool_calls` on assistant messages → Cohere `tool_calls` format (translate if spec 01 is implemented)
- `stop` → `stop_sequences`

### 3. Chat response translation

Cohere response → `llm.ChatCompletionResponse`:
```json
{"id": "...", "finish_reason": "COMPLETE", "message": {"role": "assistant", "content": [{"type": "text", "text": "Hello!"}]}, "usage": {"billed_units": {"input_tokens": 5, "output_tokens": 10}}}
```

- `finish_reason: "COMPLETE"` → `"stop"`
- `finish_reason: "MAX_TOKENS"` → `"length"`
- `finish_reason: "TOOL_CALL"` → `"tool_calls"`
- `usage.billed_units.input_tokens` → `prompt_tokens`
- `usage.billed_units.output_tokens` → `completion_tokens`

### 4. Streaming translation

Cohere streams JSON event objects (not standard SSE text/event-stream format — it uses `data: {...}\n\n` lines with typed events):
- `{"type": "content-delta", "delta": {"message": {"content": {"text": "..."}}}}` → emit content delta chunk
- `{"type": "message-end", "delta": {"finish_reason": "COMPLETE", "usage": {...}}}` → emit final chunk with usage + `[DONE]`

### 5. Embeddings translation

`llm.EmbeddingRequest` → Cohere embed request:
```json
{"model": "embed-english-v3.0", "texts": ["text to embed"], "input_type": "search_document"}
```

The `input_type` is required by Cohere but absent from the OpenAI schema. Default to `"search_document"` unless the model name contains `"search_query"` as a hint.

Cohere response → `llm.EmbeddingResponse`: extract `embeddings.float[0]` as the vector.

### 6. Configuration

Add to `pkg/config/config.go`:
```go
Cohere struct {
    APIKey string `mapstructure:"api_key"`
}
```

Register in `cmd/gateway/main.go` if `cfg.Cohere.APIKey` is non-empty. Also support DB-configured providers with `auth_type: "bearer"` and `name: "cohere"`.

## Out of Scope
- Rerank endpoint (spec 25)
- Cohere connectors / grounded generation
- Multi-step tool use (Cohere-specific feature)

## Acceptance Criteria
- [ ] Chat completion with `command-r-plus` succeeds
- [ ] Response translates `finish_reason` correctly
- [ ] Streaming works and produces OpenAI-compatible SSE chunks
- [ ] `EmbedContent` returns a valid embedding vector
- [ ] Tool calls translate correctly (if spec 01 is in place)
- [ ] Unit tests cover request/response translation with mock HTTP server

## Key Files
- `internal/proxy/provider_cohere.go` — new file
- `pkg/config/config.go` — Cohere config section
- `cmd/gateway/main.go` — register provider
