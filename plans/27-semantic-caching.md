# Spec 27: Semantic Caching

## Status
pending

## Dependencies
- Spec 19 (exact-match response caching) — `ResponseCache` interface and cache config foundation
- An embedding model must be configured in the gateway (via spec 11 generic provider or direct config)

## Goal
Cache LLM responses and retrieve them for semantically similar (not just identical) queries. Reduces cost and latency for use cases where users ask the same question in different phrasings.

## Background
Spec 19 implements exact-match caching keyed on a SHA-256 hash of the request. Semantic caching requires embedding the request's messages, performing a similarity search against cached embeddings, and returning the cached response if similarity exceeds a threshold.

Redis supports vector similarity search via the `RediSearch` module (Redis Stack). As an alternative, `qdrant` is a purpose-built vector database with a Go client.

## Requirements

### 1. Configuration

```go
SemanticCache struct {
    Enabled          bool    `mapstructure:"enabled"`
    Backend          string  `mapstructure:"backend"`           // "redis" | "qdrant"
    SimilarityThreshold float64 `mapstructure:"similarity_threshold"` // 0.0-1.0, default 0.95
    EmbeddingModel   string  `mapstructure:"embedding_model"`   // alias of an embedding-capable model
    TTL              time.Duration `mapstructure:"ttl"`         // default 24h
    MaxCachedItems   int     `mapstructure:"max_cached_items"` // soft cap, default 10000
} `mapstructure:"semantic_cache"`
```

### 2. Embedding generation

The cache generates an embedding for each incoming request by concatenating all message content and calling the configured `EmbeddingModel` via the existing `EmbeddingProvider` interface.

```go
func (sc *SemanticCache) embed(ctx context.Context, req *llm.ChatCompletionRequest) ([]float32, error) {
    text := messagesText(req.Messages) // concatenate role + content for all messages
    embReq := &llm.EmbeddingRequest{Model: sc.embeddingModel, Input: text}
    resp, err := sc.embeddingProvider.EmbedContent(ctx, embReq)
    // return resp.Data[0].Embedding
}
```

The embedding generation adds latency on the cache-miss path. This is acceptable — it's parallelizable with the actual provider call in a future optimization.

### 3. Redis backend (`internal/storage/cache_semantic_redis.go`)

Requires Redis Stack (Redis with RediSearch + RedisJSON modules). Use `redis.CreateIndex` to create a vector index:

```
FT.CREATE llmgopher:semantic:idx
  ON HASH PREFIX 1 llmgopher:semantic:
  SCHEMA
    embedding VECTOR FLAT 6 TYPE FLOAT32 DIM {dim} DISTANCE_METRIC COSINE
    response TEXT
    model TEXT
    created_at NUMERIC
```

**Cache write:**
```
HSET llmgopher:semantic:{uuid} embedding {float32_bytes} response {json} model {model} created_at {unix}
EXPIRE llmgopher:semantic:{uuid} {ttl_seconds}
```

**Cache read (KNN search):**
```
FT.SEARCH llmgopher:semantic:idx "*=>[KNN 1 @embedding $vec AS score]"
  PARAMS 2 vec {float32_bytes}
  RETURN 3 response model score
  SORTBY score ASC
```

If `score < (1 - similarity_threshold)`, return the cached response.

### 4. Qdrant backend (`internal/storage/cache_semantic_qdrant.go`)

Add `github.com/qdrant/go-client` dependency.

Create a collection `llmgopher-semantic-cache` with cosine distance. On cache write, upsert a point with the embedding vector and response JSON as payload. On cache read, query the nearest neighbor and check the score.

### 5. Integration in `internal/proxy/handler.go`

Semantic cache check runs **after** the exact-match cache check (spec 19) and **before** the provider call:

```go
// 1. Exact-match cache check (spec 19)
if cached := h.exactCache.Get(ctx, exactKey); cached != nil { ... return }

// 2. Semantic cache check
if h.semanticCache != nil {
    if cached, hit := h.semanticCache.Lookup(ctx, &req); hit {
        w.Header().Set("X-Cache", "SEMANTIC-HIT")
        // serve cached response
        return
    }
}

// 3. Provider dispatch
resp, err := provider.ChatCompletion(...)

// 4. Write to both caches
if h.exactCache != nil { h.exactCache.Set(...) }
if h.semanticCache != nil { h.semanticCache.Store(ctx, &req, resp) }
```

### 6. Cache interface extension (`pkg/llm/cache.go`)

```go
type SemanticCache interface {
    Lookup(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, bool)
    Store(ctx context.Context, req *ChatCompletionRequest, resp *ChatCompletionResponse) error
}
```

### 7. Cache response headers

- `X-Cache: SEMANTIC-HIT` — response from semantic cache
- `X-Cache-Similarity: 0.97` — similarity score of the matched entry (for debugging)

### 8. Audit logging

Add `cache_type TEXT` to `audit_log` (extends spec 19's `cache_hit` column):
- `null` — cache miss
- `'exact'` — exact match hit
- `'semantic'` — semantic hit

### 9. Admin cache management

`DELETE /v1/admin/cache` (spec 19) also clears semantic cache entries.

## Out of Scope
- Semantic cache for streaming requests
- Cache invalidation by topic/model (full flush only)
- Multiple vector stores simultaneously
- Embedding model fine-tuning

## Acceptance Criteria
- [ ] A request semantically similar to a cached request (different wording, same meaning) returns the cached response
- [ ] `X-Cache: SEMANTIC-HIT` header is present on semantic hits
- [ ] Similarity below threshold results in a cache miss (provider call)
- [ ] `X-Cache-Similarity` score is returned on hits
- [ ] Both Redis and Qdrant backends pass the same functional tests
- [ ] Embedding failures fall through to provider call (never block the request)
- [ ] Audit log correctly records `cache_type: "semantic"`

## Key Files
- `pkg/llm/cache.go` — extend with `SemanticCache` interface
- `internal/storage/cache_semantic_redis.go` — Redis Stack implementation (new file)
- `internal/storage/cache_semantic_qdrant.go` — Qdrant implementation (new file)
- `internal/proxy/handler.go` — semantic cache lookup/store
- `pkg/config/config.go` — semantic cache config
- `cmd/gateway/main.go` — init semantic cache if configured
