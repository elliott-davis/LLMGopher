package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/internal/mocks"
	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestHandler_ModelRateLimitExceeded_Returns429(t *testing.T) {
	registry := llm.NewRegistry()
	provider := &mocks.MockProvider{
		ProviderName: "openai",
		ChatResponse: &llm.ChatCompletionResponse{
			ID:      "chatcmpl-1",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o",
			Choices: []llm.Choice{
				{
					Index:   0,
					Message: &llm.Message{Role: "assistant", Content: llm.StringContent("ok")},
				},
			},
		},
	}
	registry.Register(provider, "gpt-4*")

	stateCache := buildProxyStateCache(t, 2)
	rateLimiter := middleware.NewInMemoryRateLimiter(1000, 1000)

	handler := NewHandler(
		registry,
		stateCache,
		rateLimiter,
		&mocks.MockAuditLogger{},
		&mocks.MockBudgetTracker{Remaining: 1e9},
		mocks.NewMockPricingLookup(),
		slog.Default(),
	)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want %d; body=%s", i+1, w.Code, http.StatusOK, w.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("request 3 status = %d, want %d; body=%s", w.Code, http.StatusTooManyRequests, w.Body.String())
	}
	if got := w.Header().Get("Retry-After"); got != "1" {
		t.Fatalf("Retry-After = %q, want %q", got, "1")
	}
	if !strings.Contains(w.Body.String(), "model rate limit exceeded") {
		t.Fatalf("body = %q, want model rate limit message", w.Body.String())
	}

	if len(provider.SyncCalls) != 2 {
		t.Fatalf("provider calls = %d, want 2", len(provider.SyncCalls))
	}

	allowed, err := rateLimiter.Allow(context.Background(), "key:key-001")
	if err != nil {
		t.Fatalf("key bucket check returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected key bucket to remain independent from model bucket")
	}
}

func TestHandler_ModelRateLimitZero_Disabled(t *testing.T) {
	registry := llm.NewRegistry()
	provider := &mocks.MockProvider{
		ProviderName: "openai",
		ChatResponse: &llm.ChatCompletionResponse{
			ID:      "chatcmpl-2",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o",
			Choices: []llm.Choice{
				{
					Index:   0,
					Message: &llm.Message{Role: "assistant", Content: llm.StringContent("ok")},
				},
			},
		},
	}
	registry.Register(provider, "gpt-4*")

	stateCache := buildProxyStateCache(t, 0)
	rateLimiter := middleware.NewInMemoryRateLimiter(1, 1)

	handler := NewHandler(
		registry,
		stateCache,
		rateLimiter,
		&mocks.MockAuditLogger{},
		&mocks.MockBudgetTracker{Remaining: 1e9},
		mocks.NewMockPricingLookup(),
		slog.Default(),
	)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want %d; body=%s", i+1, w.Code, http.StatusOK, w.Body.String())
		}
	}

	if len(provider.SyncCalls) != 3 {
		t.Fatalf("provider calls = %d, want 3", len(provider.SyncCalls))
	}
}

func buildProxyStateCache(t *testing.T, modelRateLimitRPS int) *storage.StateCache {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Now().UTC()
	providerID := uuid.New().String()

	mock.ExpectQuery("SELECT id, name, base_url, auth_type, has_credentials, created_at, updated_at\\s+FROM providers").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "base_url", "auth_type", "has_credentials", "created_at", "updated_at"}).
				AddRow(providerID, "openai", "https://api.openai.com/v1", "bearer", false, now, now),
		)
	mock.ExpectQuery("SELECT id, provider_id, name, alias, context_window, rate_limit_rps, created_at, updated_at\\s+FROM models").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "provider_id", "name", "alias", "context_window", "rate_limit_rps", "created_at", "updated_at"}).
				AddRow(uuid.NewString(), providerID, "gpt-4o", "gpt-4o", 128000, modelRateLimitRPS, now, now),
		)
	mock.ExpectQuery("SELECT id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, to_json\\(allowed_models\\), created_at, updated_at\\s+FROM api_keys\\s+WHERE is_active = TRUE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "key_hash", "name", "rate_limit_rps", "is_active", "expires_at", "metadata", "allowed_models", "created_at", "updated_at"}))

	cache := storage.NewStateCache(slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	cache.StartPoller(ctx, db, time.Hour)
	cancel()

	if cache.Load() == nil {
		t.Fatal("cache state is nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
	return cache
}
