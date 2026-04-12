# Spec 09: Per-Model Rate Limits

## Status
pending

## Goal
Allow individual models to have their own rate limits independent of the per-key limit. This enables operators to protect expensive models (e.g., GPT-4o) from overuse while allowing looser limits on cheaper models.

## Background
Rate limiting is enforced in `internal/middleware/ratelimit.go` by calling `RateLimiter.Allow(ctx, key)` where `key` is the API key ID from context. The `RateLimiter` interface is in `pkg/llm/ratelimiter.go`. Both `InMemoryRateLimiter` (`internal/middleware/ratelimit_memory.go`) and `RedisRateLimiter` (`internal/middleware/ratelimit_redis.go`) use a per-string-key token bucket.

The `models` table already has a `rate_limit_rps` column that is currently unused (it exists in `ModelConfig` but is never read). The state cache loads models and makes them available via `storage.StateCache`.

The request model name is not available in the rate limit middleware today — it is only known after the request body is parsed in the proxy handler.

## Requirements

### 1. Move model rate limit enforcement to the proxy handler

Model-level rate limiting cannot happen in middleware (the body isn't parsed yet). Instead, enforce it inside `internal/proxy/handler.go` after resolving the model, before dispatching to the provider.

After `h.resolveModel(req.Model)`:
```go
if err := h.checkModelRateLimit(r.Context(), modelCfg); err != nil {
    // return 429 with Retry-After header
}
```

### 2. `RateLimiter` key convention for model limits

The existing `RateLimiter.Allow(ctx, key)` signature is sufficient. Use a compound key:
```
"model:<model_alias>"
```

This keeps model and key rate limits isolated in the same underlying rate limiter without any interface changes.

### 3. `ModelConfig` RPS field

`ModelConfig.RateLimitRPS` already exists. The handler reads it from the state cache after resolving the model. If `RateLimitRPS == 0`, no model-level rate limit is enforced.

### 4. Rate limiter initialization for model buckets

Both `InMemoryRateLimiter` and `RedisRateLimiter` initialize new buckets lazily on first request — no pre-population needed. The burst for model buckets defaults to `RateLimitRPS` (same as the key-level burst).

### 5. Admin API (`internal/api/admin.go`)

`POST /v1/admin/models` and `PUT /v1/admin/models/{id}` already accept `rate_limit_rps`. No new fields needed — just ensure the value is read and the handler uses it. Verify the state cache poll query includes this column.

### 6. Response headers

When a model rate limit is hit, return:
- HTTP 429
- `Retry-After: <seconds>` header (same as the key-level rate limit does today)
- Error body: `{"error": {"type": "rate_limit_error", "message": "model rate limit exceeded"}}`

## Out of Scope
- Per-model-per-key rate limits (compound key with both model and key ID)
- Token-per-minute (TPM) rate limiting (count-based only for now)
- Metrics for model rate limit hits (covered in spec 20)

## Acceptance Criteria
- [ ] A model with `rate_limit_rps: 2` rejects a third request within the same second with 429
- [ ] A model with `rate_limit_rps: 0` has no model-level rate limit
- [ ] Model and key rate limits are independently enforced (hitting model limit doesn't affect key limit bucket)
- [ ] `Retry-After` header is present on 429 responses
- [ ] State cache change to `rate_limit_rps` takes effect within the cache poll interval (5s) without restart
- [ ] Unit test covers model rate limit enforcement in the handler

## Key Files
- `internal/proxy/handler.go` — add `checkModelRateLimit` call after model resolution
- `internal/storage/cache.go` — verify `rate_limit_rps` is included in model scan
- `internal/middleware/ratelimit_memory.go` — verify lazy bucket init works for any string key
- `internal/middleware/ratelimit_redis.go` — same
