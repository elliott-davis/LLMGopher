# Spec 18: Context-Window Fallback

## Status
pending

## Goal
When a provider rejects a request with a context-length-exceeded error, automatically retry the request with a designated larger-context model. This handles the common case where a conversation grows too long for a model's context window.

## Background
Context-length errors are distinct from general failures: they require routing to a *different, larger* model rather than retrying the same one. Spec 15 (retry) and spec 16 (fallback) handle general failures; this spec handles the specific `context_length_exceeded` error case.

Providers signal this with different error messages:
- OpenAI: `{"error": {"code": "context_length_exceeded", ...}}`
- Anthropic: `{"error": {"type": "invalid_request_error", "message": "prompt is too long: ..."}}`
- Gemini: `400` with `"Request contains content that exceeds token limit"`
- Bedrock: similar to Anthropic

## Requirements

### 1. Detection (`internal/proxy/retry.go`)

Add a helper:
```go
// isContextLengthError returns true if the error indicates the prompt exceeded the model's context window.
func isContextLengthError(err error) bool
```

Check the `ProviderError.Body` for known patterns:
- `"context_length_exceeded"` in the error code or message
- `"prompt is too long"` (Anthropic)
- `"maximum context length"` (OpenAI legacy)
- `"exceeds token limit"` (Gemini)
- `"max token limit"` (Bedrock)

This error is explicitly **not** retried by the regular retry logic in spec 15.

### 2. Model config field for context fallback

Migration (`00008_context_fallback.sql`):
```sql
ALTER TABLE models ADD COLUMN context_fallback_model TEXT;
```

Add to `ModelConfig`:
```go
ContextFallbackModel string `json:"context_fallback_model,omitempty"`
```

This is a single model alias (not a list) — the model to use when the context is too long.

### 3. Fallback in handler (`internal/proxy/handler.go`)

After a sync provider call fails:
```go
if isContextLengthError(err) && modelCfg.ContextFallbackModel != "" {
    logger.Info("context length exceeded, retrying with fallback model",
        "original_model", req.Model,
        "fallback_model", modelCfg.ContextFallbackModel,
    )
    return h.retryWithModel(ctx, req, meta, modelCfg.ContextFallbackModel)
}
```

The `retryWithModel` function resolves the fallback alias and dispatches the request to its provider. No further context fallback is attempted (only one hop).

For streaming: same detection applies — if `ChatCompletionStream` returns a context-length error before opening the stream, attempt the fallback.

### 4. Audit logging

Add `context_fallback_used bool` and include the fallback model in `fallback_used` (reusing spec 16's column). The audit log entry reflects the model that served the response.

### 5. Admin API

`POST /v1/admin/models` and `PUT /v1/admin/models/{id}` accept `context_fallback_model`. No new endpoint.

### 6. Common default patterns

Document in config comments (not enforced in code):
- `gpt-4o` → `context_fallback_model: "gpt-4o-128k"` (if the 128k variant is configured)
- `claude-3-5-sonnet` → `context_fallback_model: "claude-3-5-sonnet-200k"`

## Out of Scope
- Automatic prompt truncation (trimming oldest messages to fit in context)
- Splitting long requests across multiple calls
- Context window size tracking per request

## Acceptance Criteria
- [ ] A request to a model that returns a context-length error is retried with `context_fallback_model`
- [ ] If no `context_fallback_model` is configured, the context error is returned to the client as-is
- [ ] Context fallback is not attempted a second time (no infinite loop)
- [ ] The audit log records the fallback model
- [ ] Non-context errors are NOT routed to the context fallback model
- [ ] Detection works for OpenAI, Anthropic, and Gemini error formats

## Key Files
- `internal/storage/migrations/00008_context_fallback.sql` — new migration
- `pkg/llm/types.go` — add `ContextFallbackModel` to `ModelConfig`
- `internal/storage/cache.go` — include in model scan
- `internal/proxy/retry.go` — add `isContextLengthError`
- `internal/proxy/handler.go` — fallback dispatch on context error
- `internal/api/admin.go` — accept field in create/update
