package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ed007183/llmgopher/internal/mocks"
	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/pkg/llm"
)

var discardLogger = slog.New(slog.NewTextHandler(
	&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError + 1},
))

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// validChatBody returns a minimal valid JSON request body.
func validChatBody() string {
	return `{"model":"gpt-4o","messages":[{"role":"user","content":"hello"}]}`
}

// okHandler is a terminal handler that records whether it was reached.
func okHandler(reached *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*reached = true
		w.WriteHeader(http.StatusOK)
	})
}

// --------------------------------------------------------------------------
// Auth middleware
// --------------------------------------------------------------------------

func TestAuth_ValidKey_Passes(t *testing.T) {
	keys := map[string]string{"sk-test-abc": "key-001"}
	var reached bool

	handler := middleware.Auth(keys, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer sk-test-abc")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("handler was not reached")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_InvalidKey_Returns401(t *testing.T) {
	keys := map[string]string{"sk-test-abc": "key-001"}
	var reached bool

	handler := middleware.Auth(keys, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer sk-wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Error("handler should not be reached for invalid key")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_MissingHeader_Returns401(t *testing.T) {
	keys := map[string]string{"sk-test-abc": "key-001"}
	var reached bool

	handler := middleware.Auth(keys, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Error("handler should not be reached without auth header")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_EmptyKeyMap_AcceptsAnyToken(t *testing.T) {
	var reached bool

	handler := middleware.Auth(nil, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer any-random-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("handler should be reached with nil key map")
	}
}

func TestAuth_SetsAPIKeyIDInContext(t *testing.T) {
	keys := map[string]string{"sk-test-abc": "key-001"}
	var gotKeyID string

	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotKeyID = middleware.GetAPIKeyID(r.Context())
	})

	handler := middleware.Auth(keys, discardLogger)(inner)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer sk-test-abc")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if gotKeyID != "key-001" {
		t.Errorf("api_key_id = %q, want %q", gotKeyID, "key-001")
	}
}

// --------------------------------------------------------------------------
// RateLimit middleware
// --------------------------------------------------------------------------

func TestRateLimit_Allowed_Passes(t *testing.T) {
	limiter := &mocks.MockRateLimiter{Allowed: true}
	var reached bool

	handler := middleware.RateLimit(limiter, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(withAPIKeyID(req.Context(), "key-001"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("handler was not reached")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRateLimit_Denied_Returns429(t *testing.T) {
	limiter := &mocks.MockRateLimiter{Allowed: false}
	var reached bool

	handler := middleware.RateLimit(limiter, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(withAPIKeyID(req.Context(), "key-001"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Error("handler should not be reached when rate limited")
	}
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	if w.Header().Get("Retry-After") != "1" {
		t.Errorf("Retry-After = %q, want %q", w.Header().Get("Retry-After"), "1")
	}
}

func TestRateLimit_Error_FailsOpen(t *testing.T) {
	limiter := &mocks.MockRateLimiter{Allowed: false, Err: errors.New("redis down")}
	var reached bool

	handler := middleware.RateLimit(limiter, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(withAPIKeyID(req.Context(), "key-001"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("handler should be reached on rate limiter error (fail-open)")
	}
}

func TestRateLimit_NoKeyID_SkipsCheck(t *testing.T) {
	limiter := &mocks.MockRateLimiter{Allowed: false}
	var reached bool

	handler := middleware.RateLimit(limiter, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("handler should be reached when no key ID is present")
	}
	if len(limiter.Keys) != 0 {
		t.Errorf("limiter should not have been called, got %d calls", len(limiter.Keys))
	}
}

// --------------------------------------------------------------------------
// Guardrail middleware
// --------------------------------------------------------------------------

func TestGuardrail_Allowed_Passes(t *testing.T) {
	guard := &mocks.MockGuardrail{
		Verdict: &llm.GuardrailVerdict{Allowed: true},
	}
	var reached bool

	handler := middleware.GuardrailCheck(guard, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validChatBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("handler was not reached")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGuardrail_Blocked_Returns400(t *testing.T) {
	guard := &mocks.MockGuardrail{
		Verdict: &llm.GuardrailVerdict{Allowed: false, Reason: "toxic content detected"},
	}
	var reached bool

	handler := middleware.GuardrailCheck(guard, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validChatBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Error("handler should not be reached when guardrail blocks")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "toxic content detected") {
		t.Errorf("body should contain reason, got: %s", w.Body.String())
	}
}

func TestGuardrail_ServiceError_Returns503(t *testing.T) {
	guard := &mocks.MockGuardrail{
		Err: errors.New("guardrail timeout"),
	}
	var reached bool

	handler := middleware.GuardrailCheck(guard, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validChatBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Error("handler should not be reached on guardrail error (fail-closed)")
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestGuardrail_NilGuard_Passes(t *testing.T) {
	var reached bool

	handler := middleware.GuardrailCheck(nil, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validChatBody()))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("handler should be reached with nil guardrail")
	}
}

func TestGuardrail_InvalidJSON_Returns400(t *testing.T) {
	guard := &mocks.MockGuardrail{
		Verdict: &llm.GuardrailVerdict{Allowed: true},
	}
	var reached bool

	handler := middleware.GuardrailCheck(guard, discardLogger)(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Error("handler should not be reached with invalid JSON")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGuardrail_RestoresBody(t *testing.T) {
	guard := &mocks.MockGuardrail{
		Verdict: &llm.GuardrailVerdict{Allowed: true},
	}

	body := validChatBody()
	var innerBody string

	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		var req llm.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			innerBody = fmt.Sprintf("decode error: %v", err)
			return
		}
		innerBody = req.Model
	})

	handler := middleware.GuardrailCheck(guard, discardLogger)(inner)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if innerBody != "gpt-4o" {
		t.Errorf("inner handler got model = %q, want %q (body not restored)", innerBody, "gpt-4o")
	}
}

// --------------------------------------------------------------------------
// Full middleware chain
// --------------------------------------------------------------------------

func TestChain_OrderOfExecution(t *testing.T) {
	keys := map[string]string{"sk-test-abc": "key-001"}
	limiter := &mocks.MockRateLimiter{Allowed: true}
	guard := &mocks.MockGuardrail{
		Verdict: &llm.GuardrailVerdict{Allowed: true},
	}

	var reached bool
	inner := okHandler(&reached)

	handler := middleware.Chain(
		inner,
		middleware.Auth(keys, discardLogger),
		middleware.RateLimit(limiter, discardLogger),
		middleware.GuardrailCheck(guard, discardLogger),
	)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validChatBody()))
	req.Header.Set("Authorization", "Bearer sk-test-abc")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !reached {
		t.Error("inner handler was not reached through full chain")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if len(limiter.Keys) != 1 || limiter.Keys[0] != "key-001" {
		t.Errorf("rate limiter received keys %v, want [key-001]", limiter.Keys)
	}
	if len(guard.Calls) != 1 {
		t.Errorf("guardrail received %d calls, want 1", len(guard.Calls))
	}
}

func TestChain_AuthFailure_ShortCircuits(t *testing.T) {
	keys := map[string]string{"sk-test-abc": "key-001"}
	limiter := &mocks.MockRateLimiter{Allowed: true}
	guard := &mocks.MockGuardrail{
		Verdict: &llm.GuardrailVerdict{Allowed: true},
	}

	var reached bool
	inner := okHandler(&reached)

	handler := middleware.Chain(
		inner,
		middleware.Auth(keys, discardLogger),
		middleware.RateLimit(limiter, discardLogger),
		middleware.GuardrailCheck(guard, discardLogger),
	)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validChatBody()))
	req.Header.Set("Authorization", "Bearer wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Error("inner handler should not be reached after auth failure")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if len(limiter.Keys) != 0 {
		t.Error("rate limiter should not be called after auth failure")
	}
	if len(guard.Calls) != 0 {
		t.Error("guardrail should not be called after auth failure")
	}
}

// --------------------------------------------------------------------------
// RequestID middleware
// --------------------------------------------------------------------------

func TestRequestID_GeneratesID(t *testing.T) {
	var gotID string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotID = middleware.GetRequestID(r.Context())
	})

	handler := middleware.RequestID(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if gotID == "" {
		t.Error("request ID should be generated")
	}
	if w.Header().Get("X-Request-ID") != gotID {
		t.Errorf("response header = %q, context = %q", w.Header().Get("X-Request-ID"), gotID)
	}
}

func TestRequestID_PreservesExisting(t *testing.T) {
	var gotID string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotID = middleware.GetRequestID(r.Context())
	})

	handler := middleware.RequestID(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "custom-id-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if gotID != "custom-id-123" {
		t.Errorf("request ID = %q, want %q", gotID, "custom-id-123")
	}
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

// withAPIKeyID injects an API key ID into the context the same way Auth does.
// We use the exported GetAPIKeyID to verify, but need the unexported key constant.
// Instead, we run the Auth middleware with a matching key to set the context.
func withAPIKeyID(ctx context.Context, keyID string) context.Context {
	token := "temp-token-for-test"
	keys := map[string]string{token: keyID}

	var enrichedCtx context.Context
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		enrichedCtx = r.Context()
	})

	handler := middleware.Auth(keys, discardLogger)(inner)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	return enrichedCtx
}
