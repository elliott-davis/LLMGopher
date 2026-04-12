package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/pkg/llm"
)

// EmbeddingsHandler serves the POST /v1/embeddings endpoint.
// It resolves the provider, checks if it supports embeddings, and dispatches.
type EmbeddingsHandler struct {
	registry   llm.ProviderRegistry
	stateCache *storage.StateCache
	logger     *slog.Logger
}

func NewEmbeddingsHandler(registry llm.ProviderRegistry, logger *slog.Logger, stateCache ...*storage.StateCache) *EmbeddingsHandler {
	var cache *storage.StateCache
	if len(stateCache) > 0 {
		cache = stateCache[0]
	}

	return &EmbeddingsHandler{
		registry:   registry,
		stateCache: cache,
		logger:     logger,
	}
}

func (h *EmbeddingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req llm.EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error(), "invalid_request_error")
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "model is required", "invalid_request_error")
		return
	}
	if req.Input == "" {
		writeError(w, http.StatusBadRequest, "input is required", "invalid_request_error")
		return
	}

	resolvedModel, expectedProviderName, err := h.resolveModel(req.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}
	req.Model = resolvedModel

	var provider llm.Provider
	if expectedProviderName != "" {
		provider, err = h.registry.ResolveProvider(expectedProviderName)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err.Error(), "server_error")
			return
		}
		if !strings.EqualFold(provider.Name(), expectedProviderName) {
			writeError(w, http.StatusServiceUnavailable, "model is configured for a different provider", "server_error")
			return
		}
	} else {
		provider, err = h.registry.Resolve(req.Model)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
			return
		}
	}
	if err := enforceAllowedModel(r.Context(), h.stateCache, req.Model); err != nil {
		if errors.Is(err, errModelNotAllowed) {
			writeError(w, http.StatusForbidden, err.Error(), "permission_error")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "authorization state unavailable", "server_error")
		return
	}

	ep, ok := provider.(llm.EmbeddingProvider)
	if !ok {
		writeError(w, http.StatusNotImplemented,
			"provider "+provider.Name()+" does not support embeddings",
			"not_implemented")
		return
	}

	resp, err := ep.EmbedContent(r.Context(), &req)
	if err != nil {
		h.logger.Error("embedding error",
			"error", err,
			"provider", provider.Name(),
		)
		writeError(w, http.StatusBadGateway, "upstream provider error: "+err.Error(), "upstream_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *EmbeddingsHandler) resolveModel(requestedModel string) (resolvedModel string, expectedProviderName string, err error) {
	if h.stateCache == nil {
		return requestedModel, "", nil
	}

	state := h.stateCache.Load()
	if state == nil {
		return "", "", fmt.Errorf("routing state is not loaded yet")
	}

	modelCfg, ok := resolveConfiguredModel(state, requestedModel)
	log.Printf("embeddings routing model lookup: requested_model=%q matched_in_state_cache=%t", requestedModel, ok)
	if !ok {
		return "", "", fmt.Errorf("model %q is not configured in state cache", requestedModel)
	}

	providerID, parseErr := uuid.Parse(modelCfg.ProviderID)
	if parseErr != nil {
		return "", "", fmt.Errorf("invalid provider_id in model config")
	}

	providerCfg, ok := state.Providers[providerID]
	if !ok {
		return "", "", fmt.Errorf("no provider configured for model alias %q", requestedModel)
	}

	return modelCfg.Name, preferredProviderRegistryName(providerCfg, requestedModel), nil
}
