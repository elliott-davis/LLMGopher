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
	mu      sync.Mutex
	buckets map[string]*bucket
	rps     float64
	burst   float64
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

func NewInMemoryRateLimiter(rps, burst int) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		buckets: make(map[string]*bucket),
		rps:     float64(rps),
		burst:   float64(burst),
	}
}

func (rl *InMemoryRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.burst, lastRefill: now}
		rl.buckets[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * rl.rps
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true, nil
	}
	return false, nil
}
