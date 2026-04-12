# Spec 28: Traffic Mirroring / Shadow Mode

## Status
pending

## Goal
Send a copy of live traffic to a secondary provider asynchronously for A/B testing, cost comparison, or new model evaluation — without affecting the primary response or adding latency. This enables silent model evaluation in production conditions.

## Background
LiteLLM calls this feature "traffic mirroring" or "shadow testing". The primary provider call returns normally to the client; the mirror call runs in a fire-and-forget goroutine and its results are logged but never returned to the client.

## Requirements

### 1. Model config field

Migration (`00012_traffic_mirror.sql`):
```sql
ALTER TABLE models
    ADD COLUMN mirror_model TEXT,         -- alias of the model to mirror to
    ADD COLUMN mirror_weight INTEGER DEFAULT 100; -- % of traffic to mirror (1-100)
```

Add to `ModelConfig`:
```go
MirrorModel  string `json:"mirror_model,omitempty"`
MirrorWeight int    `json:"mirror_weight,omitempty"` // 0-100, default 100 (all traffic)
```

### 2. Mirror dispatch (`internal/proxy/mirror.go`)

```go
type MirrorWorker struct {
    registry llm.ProviderRegistry
    cache    *storage.StateCache
    logger   *slog.Logger
}

// MaybeDispatch fires a mirror request if the model has mirroring configured.
// It always returns immediately; the mirror call runs asynchronously.
func (m *MirrorWorker) MaybeDispatch(ctx context.Context, req *llm.ChatCompletionRequest, modelCfg *llm.ModelConfig)
```

Inside `MaybeDispatch`:
1. Check `modelCfg.MirrorModel != ""` and `modelCfg.MirrorWeight > 0`
2. Probabilistic sampling: if `rand.Intn(100) >= mirrorWeight`, skip
3. Fire goroutine with a detached context (not the request context — request may complete before mirror):
   ```go
   go func() {
       ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
       defer cancel()
       mirrorReq := *req // copy
       mirrorReq.Model = resolvedMirrorModel
       mirrorReq.Stream = false // mirror always sync (never stream)
       resp, err := mirrorProvider.ChatCompletion(ctx, &mirrorReq)
       // Log result — never write to client
       logger.Info("mirror response",
           "original_model", req.Model,
           "mirror_model", mirrorReq.Model,
           "latency_ms", latency,
           "error", err,
           "finish_reason", resp?.Choices[0].FinishReason,
       )
       // Write mirror audit log entry
   }()
   ```

### 3. Integration in `handler.go`

Call `MirrorWorker.MaybeDispatch` after the primary response is sent to the client:
```go
// After handleSync or handleStream completes:
if h.mirrorWorker != nil {
    h.mirrorWorker.MaybeDispatch(context.Background(), &req, modelCfg)
}
```

### 4. Mirror audit logging

Mirror calls are written to `audit_log` with a `is_mirror BOOLEAN` column:
```sql
-- In migration 00012:
ALTER TABLE audit_log ADD COLUMN is_mirror BOOLEAN NOT NULL DEFAULT FALSE;
```

Mirror entries have the same fields as regular entries. The cost is not deducted from the API key budget (mirror is for evaluation, not billable).

### 5. Mirror response comparison (optional logging)

When both the primary and mirror responses are available, log a structured comparison:
```go
logger.Info("mirror comparison",
    "request_id", meta.RequestID,
    "primary_model", req.Model,
    "primary_tokens", primaryTokens,
    "primary_cost_usd", primaryCost,
    "mirror_model", mirrorModel,
    "mirror_tokens", mirrorTokens,
    "mirror_cost_usd", mirrorCost,
    "mirror_finish_reason", resp.Choices[0].FinishReason,
)
```

### 6. Admin API

`PUT /v1/admin/models/{id}` accepts `mirror_model` and `mirror_weight`. Changes take effect within the cache poll interval.

### 7. Graceful shutdown

Mirror goroutines run with their own timeout contexts (not the request context). On gateway shutdown, in-flight mirror calls are abandoned after the graceful shutdown timeout — they are best-effort and never block shutdown.

## Out of Scope
- Automatic response comparison scoring
- Mirror results surfaced in the UI
- Streaming mirror calls
- Cost charging for mirror traffic

## Acceptance Criteria
- [ ] A model with `mirror_model: "claude-3-5-haiku"` sends a copy of each request to that model asynchronously
- [ ] The primary response is returned to the client without waiting for the mirror
- [ ] Mirror responses are logged with `is_mirror: true` in the audit log
- [ ] Mirror budget is NOT deducted from the requesting API key's budget
- [ ] `mirror_weight: 10` results in approximately 10% of requests being mirrored (probabilistic)
- [ ] Mirror errors are logged at WARN level and never bubble up to the client
- [ ] Gateway shutdown does not hang waiting for mirror goroutines

## Key Files
- `internal/storage/migrations/00012_traffic_mirror.sql` — new migration
- `pkg/llm/types.go` — `MirrorModel`, `MirrorWeight` on `ModelConfig`
- `internal/storage/cache.go` — include mirror fields in model scan
- `internal/proxy/mirror.go` — `MirrorWorker` (new file)
- `internal/proxy/handler.go` — call `MaybeDispatch` after primary dispatch
- `internal/api/admin.go` — accept mirror fields in model update
- `cmd/gateway/main.go` — initialize `MirrorWorker`
