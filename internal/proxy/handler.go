package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/pkg/llm"
)

// Handler serves the POST /v1/chat/completions endpoint.
// It parses the OpenAI-compatible request, resolves the provider, dispatches
// the call, intercepts the response for token counting, and triggers
// asynchronous cost tracking + audit logging.
type Handler struct {
	registry    llm.ProviderRegistry
	stateCache  *storage.StateCache
	rateLimiter llm.RateLimiter
	costWorker  *CostWorker
	tokenCount  *TokenCounter
	logger      *slog.Logger
}

func NewHandler(
	registry llm.ProviderRegistry,
	stateCache *storage.StateCache,
	rateLimiter llm.RateLimiter,
	auditLogger llm.AuditLogger,
	budgetTracker llm.BudgetTracker,
	pricing llm.PricingLookup,
	logger *slog.Logger,
) *Handler {
	tc := NewTokenCounter(logger)
	return &Handler{
		registry:    registry,
		stateCache:  stateCache,
		rateLimiter: rateLimiter,
		costWorker:  NewCostWorker(auditLogger, budgetTracker, pricing, tc, logger),
		tokenCount:  tc,
		logger:      logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req llm.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error(), "invalid_request_error")
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "model is required", "invalid_request_error")
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "messages must be a non-empty array", "invalid_request_error")
		return
	}

	modelCfg, expectedProviderName, err := h.resolveModel(req.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}
	req.Model = modelCfg.Name

	if err := h.checkModelRateLimit(r.Context(), modelCfg); err != nil {
		if errors.Is(err, errModelRateLimitExceeded) {
			w.Header().Set("Retry-After", "1")
			writeError(w, http.StatusTooManyRequests, "model rate limit exceeded", "rate_limit_error")
			return
		}
		h.logger.Warn("model rate limiter error, failing open",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
			"model", req.Model,
		)
	}

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

	meta := &RequestMeta{
		RequestID:    middleware.GetRequestID(r.Context()),
		APIKeyID:     middleware.GetAPIKeyID(r.Context()),
		Model:        req.Model,
		ProviderName: provider.Name(),
		Messages:     req.Messages,
		Tools:        req.Tools,
		Streaming:    req.Stream,
		StartTime:    time.Now(),
	}

	h.logger.Info("dispatching request",
		"request_id", meta.RequestID,
		"model", req.Model,
		"provider", provider.Name(),
		"stream", req.Stream,
	)

	if req.Stream {
		h.handleStream(w, r, provider, &req, meta)
	} else {
		h.handleSync(w, r, provider, &req, meta)
	}
}

func (h *Handler) resolveModel(requestedModel string) (modelCfg *llm.Model, expectedProviderName string, err error) {
	if h.stateCache == nil {
		return &llm.Model{
			Name:  requestedModel,
			Alias: requestedModel,
		}, "", nil
	}

	state := h.stateCache.Load()
	if state == nil {
		return nil, "", fmt.Errorf("routing state is not loaded yet")
	}

	modelCfg, ok := resolveConfiguredModel(state, requestedModel)
	log.Printf("chat routing model lookup: requested_model=%q matched_in_state_cache=%t", requestedModel, ok)
	if !ok {
		return nil, "", fmt.Errorf("model %q is not configured in state cache", requestedModel)
	}

	providerID, parseErr := uuid.Parse(modelCfg.ProviderID)
	if parseErr != nil {
		return nil, "", fmt.Errorf("invalid provider_id in model config")
	}

	providerCfg, ok := state.Providers[providerID]
	if !ok {
		return nil, "", fmt.Errorf("no provider configured for model alias %q", requestedModel)
	}

	return modelCfg, preferredProviderRegistryName(providerCfg, requestedModel), nil
}

var errModelRateLimitExceeded = errors.New("model rate limit exceeded")

func (h *Handler) checkModelRateLimit(ctx context.Context, modelCfg *llm.Model) error {
	if h.rateLimiter == nil || modelCfg == nil || modelCfg.RateLimitRPS <= 0 {
		return nil
	}

	key := "model:" + modelCfg.Alias
	if modelCfg.Alias == "" {
		key = "model:" + modelCfg.Name
	}

	if configurable, ok := h.rateLimiter.(llm.ConfigurableRateLimiter); ok {
		configurable.SetLimit(key, modelCfg.RateLimitRPS)
	}

	allowed, err := h.rateLimiter.Allow(ctx, key)
	if err != nil {
		return err
	}
	if !allowed {
		return errModelRateLimitExceeded
	}
	return nil
}

func (h *Handler) handleSync(w http.ResponseWriter, r *http.Request, provider llm.Provider, req *llm.ChatCompletionRequest, meta *RequestMeta) {
	resp, err := provider.ChatCompletion(r.Context(), req)
	if err != nil {
		h.logger.Error("provider error",
			"error", err,
			"provider", provider.Name(),
			"request_id", meta.RequestID,
		)
		h.costWorker.RecordError(meta, http.StatusBadGateway, err.Error())
		writeError(w, http.StatusBadGateway, "upstream provider error: "+err.Error(), "upstream_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

	// Fire-and-forget: count tokens, calculate cost, deduct budget, write audit log.
	h.costWorker.RecordSync(meta, resp, http.StatusOK)
}

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request, provider llm.Provider, req *llm.ChatCompletionRequest, meta *RequestMeta) {
	stream, err := provider.ChatCompletionStream(r.Context(), req)
	if err != nil {
		h.logger.Error("provider stream error",
			"error", err,
			"provider", provider.Name(),
			"request_id", meta.RequestID,
		)
		h.costWorker.RecordError(meta, http.StatusBadGateway, err.Error())
		writeError(w, http.StatusBadGateway, "upstream provider error: "+err.Error(), "upstream_error")
		return
	}

	// Wrap the raw stream with an interceptor that captures text deltas
	// and counts tokens on-the-fly without adding latency to the client.
	interceptor := NewStreamInterceptor(stream, req.Model, h.tokenCount)

	flusher, ok := w.(http.Flusher)
	if !ok {
		interceptor.Close()
		writeError(w, http.StatusInternalServerError, "streaming not supported", "server_error")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Relay intercepted SSE data to the client, flushing after each read.
	buf := make([]byte, 4096)
	for {
		n, readErr := interceptor.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
		}
		if readErr != nil {
			if readErr != io.EOF {
				h.logger.Error("stream read error",
					"error", readErr,
					"request_id", meta.RequestID,
				)
			}
			break
		}
	}

	// Fire-and-forget: the cost worker waits for the interceptor to finish
	// accumulating, then counts tokens, calculates cost, deducts budget,
	// and writes the audit log — all asynchronously.
	h.costWorker.RecordStream(meta, interceptor, http.StatusOK)
}

func writeError(w http.ResponseWriter, status int, msg, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(llm.APIError{
		Error: llm.APIErrorBody{
			Message: msg,
			Type:    errType,
		},
	})
}
