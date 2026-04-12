package middleware

import (
	"context"
	"sync"
	"time"
)

// InMemoryRateLimiter is a simple token bucket rate limiter that runs
// entirely in-process. Used as the fallback when Redis is not configured.
// Not suitable for multi-instance deployments.
type InMemoryRateLimiter struct {
	mu           sync.Mutex
	buckets      map[string]*bucket
	limits       map[string]bucketLimit
	defaultLimit bucketLimit
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type bucketLimit struct {
	rps   float64
	burst float64
}

func NewInMemoryRateLimiter(rps, burst int) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		buckets: make(map[string]*bucket),
		limits:  make(map[string]bucketLimit),
		defaultLimit: bucketLimit{
			rps:   float64(rps),
			burst: float64(burst),
		},
	}
}

func (rl *InMemoryRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit := rl.limitForKey(key)
	if limit.rps <= 0 || limit.burst <= 0 {
		return true, nil
	}

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: limit.burst, lastRefill: now}
		rl.buckets[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * limit.rps
	if b.tokens > limit.burst {
		b.tokens = limit.burst
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true, nil
	}
	return false, nil
}

func (rl *InMemoryRateLimiter) SetLimit(key string, rps int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rps <= 0 {
		delete(rl.limits, key)
		return
	}
	rl.limits[key] = bucketLimit{rps: float64(rps), burst: float64(rps)}
}

func (rl *InMemoryRateLimiter) limitForKey(key string) bucketLimit {
	if limit, ok := rl.limits[key]; ok {
		return limit
	}
	return rl.defaultLimit
}
