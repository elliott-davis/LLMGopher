# Spec 15: Provider Retry with Exponential Backoff

## Status
pending

## Goal
Automatically retry failed provider requests on transient errors (rate limits, server errors, timeouts) with exponential backoff. This dramatically improves reliability without requiring clients to implement retry logic.

## Background
`internal/proxy/handler.go` calls `provider.ChatCompletion()` or `provider.ChatCompletionStream()` once and returns any error to the client immediately. There is no retry logic.

`github.com/sethvargo/go-retry` is already in `go.sum` as an indirect dependency — it can be promoted to a direct dependency without adding a new module.

Streaming requests cannot be retried once the response has been started (headers sent). Only pre-response failures are retryable.

## Requirements

### 1. Retry configuration

Add to `pkg/config/config.go` under the `Gateway` section:
```go
Retry struct {
    MaxAttempts int           `mapstructure:"max_attempts"`  // default 3
    InitialWait time.Duration `mapstructure:"initial_wait"`  // default 500ms
    MaxWait     time.Duration `mapstructure:"max_wait"`      // default 10s
} `mapstructure:"retry"`
```

Env vars: `LLMGOPHER_RETRY_MAX_ATTEMPTS`, etc.

### 2. Retryable error detection

Create `internal/proxy/retry.go`:
```go
// isRetryable returns true if the error or HTTP status warrants a retry.
func isRetryable(statusCode int, err error) bool
```

Retry on:
- HTTP 429 (rate limit) — always retry, use `Retry-After` header if present
- HTTP 500, 502, 503, 504 (server errors)
- Connection errors, timeouts (`context.DeadlineExceeded` is NOT retried — that means the client's deadline was hit)
- `io.ErrUnexpectedEOF` (connection dropped)

Do NOT retry on:
- HTTP 400, 401, 403, 404 (client errors)
- `context.Canceled` (client disconnected)
- Any error from the context itself

### 3. Retry wrapper for sync requests (`internal/proxy/handler.go`)

Wrap the `provider.ChatCompletion()` call:

```go
var resp *llm.ChatCompletionResponse
err = retry.Do(ctx, retry.WithMaxRetries(cfg.MaxAttempts-1,
    retry.NewExponential(cfg.InitialWait)), func(ctx context.Context) error {
    var callErr error
    resp, callErr = provider.ChatCompletion(ctx, req)
    if callErr != nil {
        if isRetryable(0, callErr) {
            return retry.RetryableError(callErr)
        }
        return callErr // non-retryable, stop immediately
    }
    return nil
})
```

For HTTP status errors: the provider returns an error that wraps the status code. Extract the status via a type assertion to `*ProviderError` (see requirement 5).

### 4. Streaming retry

Retry is only applied before the stream is opened. Once `provider.ChatCompletionStream()` returns a non-nil `io.ReadCloser`, the response headers are committed — no retry is possible. The retry wrapper only applies to the initial `ChatCompletionStream()` call, not to reads from the stream.

### 5. Structured provider errors

Introduce `ProviderError` in `internal/proxy/`:
```go
type ProviderError struct {
    StatusCode int
    Body       string
    RetryAfter time.Duration // parsed from Retry-After header, 0 if absent
}
func (e *ProviderError) Error() string
```

All provider implementations should return this type for HTTP error responses instead of `fmt.Errorf`. This allows the retry logic to inspect the status code without string parsing.

Update `provider_openai.go`, `provider_anthropic.go`, `provider_openai_compat.go`, and `provider_bedrock.go` to return `*ProviderError` on non-2xx responses.

### 6. Retry-After header handling

When a `ProviderError` has `RetryAfter > 0`, use it as the backoff duration for that attempt instead of the exponential backoff:
```go
if pe.RetryAfter > 0 {
    time.Sleep(pe.RetryAfter)
    return retry.RetryableError(pe)
}
```

### 7. Logging

Log each retry attempt:
```go
logger.Warn("provider request failed, retrying",
    "attempt", attemptN,
    "max_attempts", maxAttempts,
    "error", err,
    "provider", providerName,
    "request_id", requestID,
)
```

### 8. Audit logging on retry

The audit log entry should reflect the final status (success or final failure). Add `retry_count int` to `llm.AuditEntry` and the `audit_log` schema (migration `00005_audit_retry_count.sql`).

## Out of Scope
- Circuit breaker (spec 16 builds on this)
- Per-provider retry config (global config is sufficient for now)
- Retry for streaming mid-stream errors

## Acceptance Criteria
- [ ] A 429 from a provider is retried up to `max_attempts` times
- [ ] A 400 from a provider is NOT retried
- [ ] `context.Canceled` is NOT retried
- [ ] Exponential backoff increases delay between attempts
- [ ] `Retry-After` header value is respected
- [ ] Each retry attempt is logged with attempt number
- [ ] Final failure returns the last error to the client
- [ ] `retry_count` is recorded in the audit log
- [ ] Unit tests cover retryable vs non-retryable error detection

## Key Files
- `internal/proxy/retry.go` — new file: `isRetryable`, `ProviderError`
- `internal/proxy/handler.go` — wrap sync and stream calls
- `pkg/config/config.go` — retry config
- `internal/storage/migrations/00005_audit_retry_count.sql` — new migration
- `pkg/llm/audit.go` — add `RetryCount` field
