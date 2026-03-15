package middleware

import (
	"log/slog"
	"net/http"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// RateLimit enforces per-key request throttling using the RateLimiter interface.
func RateLimit(limiter llm.RateLimiter, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := GetAPIKeyID(r.Context())
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			allowed, err := limiter.Allow(r.Context(), key)
			if err != nil {
				// Fail open on infrastructure errors.
				logger.Warn("rate limiter error, failing open",
					"error", err,
					"request_id", GetRequestID(r.Context()),
				)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				logger.Info("rate limited",
					"key", key,
					"request_id", GetRequestID(r.Context()),
				)
				w.Header().Set("Retry-After", "1")
				http.Error(w,
					`{"error":{"message":"rate limit exceeded","type":"rate_limit_error"}}`,
					http.StatusTooManyRequests,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
