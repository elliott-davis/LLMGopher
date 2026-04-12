# Spec 19: Exact-Match Response Caching

## Status
pending

## Goal
Cache completed (non-streaming) LLM responses and return them immediately on identical requests. Reduces latency to near-zero and eliminates cost for repeated identical prompts — common in testing, evals, and demo scenarios.

## Background
Redis is already wired for rate limiting (`internal/middleware/ratelimit_redis.go`). The `CacheConfig` in `pkg/config/config.go` has Redis connection details. Non-streaming responses are fully materialized in `internal/proxy/handler.go` before being written to the client.

## Requirements

### 1. Cache key

The cache key is a SHA-256 hash of the deterministic fields of the request:
```go
func cacheKey(req *llm.ChatCompletionRequest) string {
    h := sha256.New()
    // Encode only deterministic fields — exclude user, stream (always false for cached)
    fields := struct {
        Model            string
        Messages         []llm.Message
        Temperature      *float64
        TopP             *float64
        MaxTokens        *int
        Stop             json.RawMessage
        PresencePenalty  float64
        FrequencyPenalty float64
        Tools            []llm.Tool   // if spec 01 is in place
    }{...}
    json.NewEncoder(h).Encode(fields)
    return "llmgopher:cache:" + hex.EncodeToString(h.Sum(nil))
}
```

Exclude: `user`, `stream`, `n` (irrelevant for caching purposes), `metadata`.

### 2. Cache config

Add to `pkg/config/config.go`:
```go
Cache struct {
    Enabled    bool          `mapstructure:"enabled"`       // default false
    TTL        time.Duration `mapstructure:"ttl"`           // default 1h
    MaxSizeKB  int           `mapstructure:"max_size_kb"`   // max cached response size, default 512KB
} `mapstructure:"cache"`
```

Env: `LLMGOPHER_CACHE_ENABLED`, `LLMGOPHER_CACHE_TTL`, `LLMGOPHER_CACHE_MAX_SIZE_KB`.

Cache is disabled by default. Requires Redis to be enabled.

### 3. Cache middleware / integration point

Add cache check and write in `internal/proxy/handler.go`, in `handleSync()`:

**Before dispatch:**
```go
if h.cache != nil {
    key := cacheKey(&req)
    if cached := h.cache.Get(ctx, key); cached != nil {
        // Deserialize and write cached response
        w.Header().Set("X-Cache", "HIT")
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write(cached)
        return
    }
}
```

**After successful response:**
```go
if h.cache != nil && len(respBytes) < cfg.MaxSizeKB*1024 {
    key := cacheKey(&req)
    h.cache.Set(ctx, key, respBytes, cfg.TTL)
}
```

### 4. Cache interface (`pkg/llm/cache.go`)

```go
type ResponseCache interface {
    Get(ctx context.Context, key string) []byte  // nil = miss
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

### 5. Redis implementation (`internal/storage/cache_response.go`)

```go
type RedisResponseCache struct {
    client *redis.Client
}

func NewRedisResponseCache(client *redis.Client) *RedisResponseCache
func (c *RedisResponseCache) Get(ctx context.Context, key string) []byte
func (c *RedisResponseCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
func (c *RedisResponseCache) Delete(ctx context.Context, key string) error
```

### 6. Streaming exclusion

Cache only applies to non-streaming requests (`req.Stream == false`). Streaming responses are never cached or served from cache.

### 7. Per-request cache bypass

Support `Cache-Control: no-cache` request header to bypass the cache (read fresh from provider but do not update cache). Support `Cache-Control: no-store` to bypass both read and write.

### 8. Cache headers in response

- `X-Cache: HIT` — response served from cache
- `X-Cache: MISS` — response fetched from provider
- `X-Cache-Key: <hash>` — the cache key (for debugging)

### 9. Audit logging

Cache hits should still be logged to the audit log with `status_code: 200` but `cost_usd: 0` and `latency_ms` reflecting the cache retrieval time (not provider latency). Add a `cache_hit bool` column:

```sql
-- migration 00009_audit_cache.sql
ALTER TABLE audit_log ADD COLUMN cache_hit BOOLEAN NOT NULL DEFAULT FALSE;
```

### 10. Admin cache invalidation

`DELETE /v1/admin/cache` — flush the entire cache namespace (`SCAN` + `DEL` on `llmgopher:cache:*` keys). Returns count of deleted keys.

## Out of Scope
- Semantic caching (spec 27)
- In-memory cache (Redis is required)
- Cache invalidation by model or key (full flush only for now)
- Caching streaming responses

## Acceptance Criteria
- [ ] Identical non-streaming requests return the cached response on the second call
- [ ] `X-Cache: HIT` header present on cached responses
- [ ] `X-Cache: MISS` on first call and on `Cache-Control: no-cache`
- [ ] Cache is bypassed for streaming requests
- [ ] TTL is respected (response not returned after expiry)
- [ ] Responses larger than `MaxSizeKB` are not cached
- [ ] Audit log records `cache_hit: true` and `cost_usd: 0` for cache hits
- [ ] `DELETE /v1/admin/cache` clears all cache entries and returns count

## Key Files
- `pkg/llm/cache.go` — `ResponseCache` interface (new file)
- `internal/storage/cache_response.go` — Redis implementation (new file)
- `internal/proxy/handler.go` — cache check/write in `handleSync`
- `internal/proxy/cache_key.go` — cache key hashing (new file)
- `pkg/config/config.go` — cache config
- `cmd/gateway/main.go` — initialize cache if enabled
- `internal/storage/migrations/00009_audit_cache.sql` — new migration
- `internal/api/admin.go` — cache flush endpoint
