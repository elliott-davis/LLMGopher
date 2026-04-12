# Spec 16: Provider Fallback Chains

## Status
pending

## Goal
When a provider exhausts all retries and still fails, automatically retry the request against a configured fallback provider or model. This enables zero-client-change resilience across provider outages and capacity issues.

## Background
Spec 15 adds per-provider retry with backoff. This spec adds a second layer: when all retries are exhausted, the request is tried on the next provider in a fallback list.

`ModelConfig` currently has: `id`, `provider_id`, `name`, `alias`, `context_window`. Fallback configuration needs to be added to the model config.

## Requirements

### 1. Migration (`internal/storage/migrations/00006_model_fallbacks.sql`)

```sql
ALTER TABLE models ADD COLUMN fallback_models TEXT[];
```

`fallback_models` is an ordered list of model aliases to try if the primary fails. Example: `["gpt-4o-mini", "claude-3-haiku"]`.

### 2. Update `ModelConfig` (`pkg/llm/types.go`)

```go
FallbackModels []string `json:"fallback_models,omitempty"`
```

Include in state cache load query and scan.

### 3. Fallback loop in `internal/proxy/handler.go`

After all retries are exhausted for the primary provider:

```go
func (h *Handler) dispatchWithFallback(ctx context.Context, req *llm.ChatCompletionRequest, meta *RequestMeta, fallbacks []string) (*llm.ChatCompletionResponse, error) {
    // Try each fallback in order
    for _, fallbackAlias := range fallbacks {
        fallbackModel, fallbackProvider, err := h.resolveModel(fallbackAlias)
        if err != nil {
            h.logger.Warn("fallback model not resolvable", "alias", fallbackAlias, "error", err)
            continue
        }
        fallbackReq := *req // copy
        fallbackReq.Model = fallbackModel
        resp, err := h.tryWithRetry(ctx, fallbackProvider, &fallbackReq)
        if err == nil {
            meta.FallbackUsed = fallbackAlias
            return resp, nil
        }
        h.logger.Warn("fallback provider failed", "alias", fallbackAlias, "error", err)
    }
    return nil, fmt.Errorf("all providers failed including %d fallbacks", len(fallbacks))
}
```

### 4. Streaming fallback

For streaming requests, fallback is only possible before the response stream is opened. The same `dispatchWithFallback` logic applies, but using `ChatCompletionStream` instead of `ChatCompletion`. Once a stream is open, no fallback is attempted.

### 5. Audit logging

Add `FallbackModel string` to `llm.AuditEntry`. Record which model/provider actually served the request. The `model` and `provider` fields in the audit log should reflect the model that successfully responded, not the originally requested model. Add an `original_model` column.

Migration `00006_model_fallbacks.sql` also:
```sql
ALTER TABLE audit_log
    ADD COLUMN original_model TEXT,
    ADD COLUMN fallback_used TEXT;
```

### 6. Admin API

`POST /v1/admin/models` and `PUT /v1/admin/models/{id}` accept `fallback_models: ["alias1", "alias2"]`. No new endpoint needed.

### 7. Error response when all fallbacks fail

Return the original provider's error to the client, not a generic fallback error. Include an `X-LLMGopher-Fallbacks-Tried` header listing the attempted fallback aliases (for debugging).

## Out of Scope
- Content-policy fallbacks (routing based on the type of error rather than exhausted retries)
- Context-window fallback (spec 18 handles the specific case of `context_length_exceeded`)
- Global fallback (a catch-all for all models) — use per-model config

## Acceptance Criteria
- [ ] When the primary provider returns 5xx errors, the first fallback in `fallback_models` is tried
- [ ] Successful fallback is reflected in the audit log (`fallback_used`, `original_model`)
- [ ] When all fallbacks fail, the response returns the original error with `X-LLMGopher-Fallbacks-Tried` header
- [ ] Model with no `fallback_models` behaves exactly as before (no regression)
- [ ] Fallback is not attempted on 4xx client errors
- [ ] Migration adds new columns cleanly
- [ ] `PUT /v1/admin/models/{id}` saves `fallback_models`

## Key Files
- `internal/storage/migrations/00006_model_fallbacks.sql` — new migration
- `pkg/llm/types.go` — `FallbackModels` on `ModelConfig`, `FallbackModel`/`OriginalModel` on `AuditEntry`
- `internal/storage/cache.go` — include `fallback_models` in model scan
- `internal/proxy/handler.go` — fallback loop in sync and stream dispatch
- `internal/api/admin.go` — accept `fallback_models` in create/update
