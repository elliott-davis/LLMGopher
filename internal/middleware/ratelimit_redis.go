package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements a token bucket rate limiter backed by Redis.
// Uses a Lua script for atomic check-and-decrement to avoid race conditions.
type RedisRateLimiter struct {
	client       *redis.Client
	defaultLimit bucketLimit
	mu           sync.RWMutex
	limits       map[string]bucketLimit
	logger       *slog.Logger
}

func NewRedisRateLimiter(client *redis.Client, rps, burst int, logger *slog.Logger) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		defaultLimit: bucketLimit{
			rps:   float64(rps),
			burst: float64(burst),
		},
		limits: make(map[string]bucketLimit),
		logger: logger,
	}
}

// tokenBucketScript atomically refills and consumes from a token bucket.
// KEYS[1] = bucket key
// ARGV[1] = max tokens (burst)
// ARGV[2] = refill rate (tokens per second)
// ARGV[3] = current timestamp in microseconds
// Returns 1 if allowed, 0 if denied.
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local max_tokens = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

if tokens == nil then
    tokens = max_tokens
    last_refill = now
end

local elapsed = (now - last_refill) / 1000000.0
local refill = elapsed * refill_rate
tokens = math.min(max_tokens, tokens + refill)

if tokens >= 1 then
    tokens = tokens - 1
    redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
    redis.call("EXPIRE", key, math.ceil(max_tokens / refill_rate) + 1)
    return 1
else
    redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
    redis.call("EXPIRE", key, math.ceil(max_tokens / refill_rate) + 1)
    return 0
end
`)

func (rl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	limit := rl.limitForKey(key)
	if limit.rps <= 0 || limit.burst <= 0 {
		return true, nil
	}

	bucketKey := fmt.Sprintf("rl:%s", key)
	nowMicro := time.Now().UnixMicro()

	result, err := tokenBucketScript.Run(ctx, rl.client, []string{bucketKey},
		limit.burst, limit.rps, nowMicro,
	).Int()

	if err != nil {
		return false, fmt.Errorf("redis rate limit script: %w", err)
	}

	return result == 1, nil
}

func (rl *RedisRateLimiter) SetLimit(key string, rps int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rps <= 0 {
		delete(rl.limits, key)
		return
	}
	rl.limits[key] = bucketLimit{rps: float64(rps), burst: float64(rps)}
}

func (rl *RedisRateLimiter) limitForKey(key string) bucketLimit {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if limit, ok := rl.limits[key]; ok {
		return limit
	}
	return rl.defaultLimit
}
