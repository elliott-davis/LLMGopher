package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ed007183/llmgopher/internal/storage"
)

const apiKeyIDKey ctxKey = 1

// Auth validates the Bearer token in the Authorization header against
// configured API keys. The resolved key ID is stored in the context for
// downstream rate limiting and audit logging.
//
// apiKeys maps raw API key strings to their logical key IDs.
// If apiKeys is nil or empty, any non-empty bearer token is accepted
// (useful for development).
func Auth(apiKeys map[string]string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				logger.Warn("missing or malformed authorization header",
					"request_id", GetRequestID(r.Context()),
				)
				http.Error(w,
					`{"error":{"message":"missing or invalid API key","type":"authentication_error"}}`,
					http.StatusUnauthorized,
				)
				return
			}

			var keyID string
			if len(apiKeys) > 0 {
				var ok bool
				keyID, ok = apiKeys[token]
				if !ok {
					logger.Warn("invalid API key",
						"request_id", GetRequestID(r.Context()),
					)
					http.Error(w,
						`{"error":{"message":"invalid API key","type":"authentication_error"}}`,
						http.StatusUnauthorized,
					)
					return
				}
			} else {
				keyID = token
			}

			ctx := context.WithValue(r.Context(), apiKeyIDKey, keyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthWithStateCache validates bearer tokens against the hot-swapped
// in-memory state cache. Tokens are SHA-256 hashed before lookup, matching
// api_keys.key_hash values stored in the database.
func AuthWithStateCache(stateCache *storage.StateCache, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				logger.Warn("missing or malformed authorization header",
					"request_id", GetRequestID(r.Context()),
				)
				http.Error(w,
					`{"error":{"message":"missing or invalid API key","type":"authentication_error"}}`,
					http.StatusUnauthorized,
				)
				return
			}

			if stateCache == nil {
				http.Error(w,
					`{"error":{"message":"authentication service unavailable","type":"server_error"}}`,
					http.StatusServiceUnavailable,
				)
				return
			}

			state := stateCache.Load()
			if state == nil {
				logger.Warn("gateway auth state unavailable",
					"request_id", GetRequestID(r.Context()),
				)
				http.Error(w,
					`{"error":{"message":"authentication service unavailable","type":"server_error"}}`,
					http.StatusServiceUnavailable,
				)
				return
			}

			hashedKey := hashAPIKey(token)
			key, ok := state.APIKeys[hashedKey]
			if !ok {
				logger.Warn("invalid API key",
					"request_id", GetRequestID(r.Context()),
				)
				http.Error(w,
					`{"error":{"message":"invalid API key","type":"authentication_error"}}`,
					http.StatusUnauthorized,
				)
				return
			}

			ctx := context.WithValue(r.Context(), apiKeyIDKey, key.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAPIKeyID extracts the authenticated API key ID from the context.
func GetAPIKeyID(ctx context.Context) string {
	if id, ok := ctx.Value(apiKeyIDKey).(string); ok {
		return id
	}
	return ""
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func hashAPIKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}
