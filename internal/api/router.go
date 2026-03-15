package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/internal/proxy"
	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/internal/validation"
	"github.com/ed007183/llmgopher/pkg/llm"
)

// Dependencies groups the services required by the API handlers.
type Dependencies struct {
	Registry               llm.ProviderRegistry
	RateLimiter            llm.RateLimiter
	Guardrail              llm.Guardrail
	AuditLogger            llm.AuditLogger
	BudgetTracker          llm.BudgetTracker
	Pricing                llm.PricingLookup
	StateCache             *storage.StateCache
	DB                     *sql.DB
	CredentialValidator    *validation.CredentialValidator
	ProviderCredentialsKey []byte
	APIKeys                map[string]string // raw key -> key_id; nil = accept any
	Logger                 *slog.Logger
}

// NewRouter builds the HTTP handler tree with middleware and routes.
// Uses Go 1.22+ enhanced routing patterns (method + path).
func NewRouter(deps *Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /ready", handleReady)
	mux.HandleFunc("GET /v1/admin/keys", HandleGetKeys(deps.StateCache))
	mux.HandleFunc("GET /v1/admin/models", HandleGetModels(deps.StateCache))
	mux.HandleFunc("GET /v1/admin/providers", HandleGetProviders(deps.StateCache))
	mux.HandleFunc("POST /v1/admin/keys", HandleCreateAPIKey(deps.DB))
	mux.HandleFunc("POST /api/v1/credentials/validate", HandleValidateCredential(deps.CredentialValidator))
	mux.HandleFunc("POST /v1/admin/providers", HandleCreateProviderWithValidation(deps.DB, deps.ProviderCredentialsKey, deps.CredentialValidator))
	mux.HandleFunc("POST /v1/admin/models", HandleCreateModel(deps.DB))
	mux.HandleFunc("PUT /v1/admin/providers/{id}", HandleUpdateProviderWithValidation(deps.DB, deps.ProviderCredentialsKey, deps.CredentialValidator))
	mux.HandleFunc("DELETE /v1/admin/providers/{id}", HandleDeleteProvider(deps.DB))
	mux.HandleFunc("PUT /v1/admin/models/{id}", HandleUpdateModel(deps.DB))
	mux.HandleFunc("DELETE /v1/admin/models/{id}", HandleDeleteModel(deps.DB))

	chatHandler := proxy.NewHandler(deps.Registry, deps.StateCache, deps.AuditLogger, deps.BudgetTracker, deps.Pricing, deps.Logger)
	embeddingsHandler := proxy.NewEmbeddingsHandler(deps.Registry, deps.Logger, deps.StateCache)

	// OpenAI-compatible endpoints — protected by the full middleware chain.
	mux.Handle("POST /v1/chat/completions", applyChatMiddleware(deps, chatHandler))
	mux.Handle("POST /v1/embeddings", applyChatMiddleware(deps, embeddingsHandler))

	// Wrap the entire mux with top-level middleware (logging, recovery, request ID).
	handler := middleware.Chain(
		mux,
		middleware.Recovery(deps.Logger),
		middleware.RequestID,
		middleware.Logging(deps.Logger),
	)

	return handler
}

// applyChatMiddleware layers auth, rate limiting, and guardrails onto a handler.
func applyChatMiddleware(deps *Dependencies, next http.Handler) http.Handler {
	authMW := middleware.Auth(deps.APIKeys, deps.Logger)
	if deps.StateCache != nil {
		authMW = middleware.AuthWithStateCache(deps.StateCache, deps.Logger)
	}

	return middleware.Chain(
		next,
		authMW,
		middleware.RateLimit(deps.RateLimiter, deps.Logger),
		middleware.GuardrailCheck(deps.Guardrail, deps.Logger),
	)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleReady(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
