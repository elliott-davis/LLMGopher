package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ed007183/llmgopher/internal/mocks"
	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestEmbeddingsHandler_NonEmbeddingProvider_Returns501(t *testing.T) {
	registry := llm.NewRegistry()
	provider := &mocks.MockProvider{ProviderName: "openai"}
	registry.Register(provider, "gpt-4*")

	handler := NewEmbeddingsHandler(registry, slog.Default())

	body := `{"model":"gpt-4o","input":"hello world"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
	if !strings.Contains(w.Body.String(), "does not support embeddings") {
		t.Errorf("body = %q, want 'does not support embeddings'", w.Body.String())
	}
}

func TestEmbeddingsHandler_MissingModel(t *testing.T) {
	registry := llm.NewRegistry()
	handler := NewEmbeddingsHandler(registry, slog.Default())

	body := `{"input":"hello world"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestEmbeddingsHandler_MissingInput(t *testing.T) {
	registry := llm.NewRegistry()
	provider := &mocks.MockProvider{ProviderName: "openai"}
	registry.Register(provider, "gpt-4*")

	handler := NewEmbeddingsHandler(registry, slog.Default())

	body := `{"model":"gpt-4o"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestEmbeddingsHandler_InvalidJSON(t *testing.T) {
	registry := llm.NewRegistry()
	handler := NewEmbeddingsHandler(registry, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", strings.NewReader("{bad"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestEmbeddingsHandler_UnknownModel(t *testing.T) {
	registry := llm.NewRegistry()
	handler := NewEmbeddingsHandler(registry, slog.Default())

	body := `{"model":"unknown-model","input":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
