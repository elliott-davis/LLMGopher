package llm

import "context"

// RateLimiter controls request throughput per identity (API key, user, org).
// Implementations may use Redis token buckets, in-memory counters, or a no-op.
type RateLimiter interface {
	// Allow checks whether the given key is permitted to proceed.
	// It returns true if the request is allowed, false if rate-limited.
	// The returned error signals infrastructure failures (e.g. Redis down);
	// callers should fail-open on error.
	Allow(ctx context.Context, key string) (bool, error)
}

// ConfigurableRateLimiter extends RateLimiter with per-key limit overrides.
// Implementations may ignore overrides for keys not explicitly configured.
type ConfigurableRateLimiter interface {
	RateLimiter
	SetLimit(key string, rps int)
}
