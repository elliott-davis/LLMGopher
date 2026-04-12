package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ed007183/llmgopher/internal/api"
	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/google/uuid"
)

type modelsListResponse struct {
	Object string            `json:"object"`
	Data   []modelsListEntry `json:"data"`
}

type modelsListEntry struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

func TestHandleListModels_EmptyCache_ReturnsEmptyList(t *testing.T) {
	handler := api.HandleListModels(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body modelsListResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Object != "list" {
		t.Fatalf("object = %q, want %q", body.Object, "list")
	}
	if len(body.Data) != 0 {
		t.Fatalf("data length = %d, want 0", len(body.Data))
	}
}

func TestHandleListModels_ReturnsOpenAICompatiblePayload(t *testing.T) {
	cache := buildModelStateCache(t)
	handler := api.HandleListModels(cache)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body modelsListResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Object != "list" {
		t.Fatalf("object = %q, want %q", body.Object, "list")
	}
	if len(body.Data) != 2 {
		t.Fatalf("data length = %d, want 2", len(body.Data))
	}

	if body.Data[0].ID != "alpha-model" {
		t.Fatalf("data[0].id = %q, want %q", body.Data[0].ID, "alpha-model")
	}
	if body.Data[1].ID != "zeta-model" {
		t.Fatalf("data[1].id = %q, want %q", body.Data[1].ID, "zeta-model")
	}

	expectedCreated := map[string]int64{
		"alpha-model": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		"zeta-model":  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC).Unix(),
	}
	for _, item := range body.Data {
		if item.Object != "model" {
			t.Fatalf("entry object = %q, want %q", item.Object, "model")
		}
		if item.OwnedBy != "llmgopher" {
			t.Fatalf("owned_by = %q, want %q", item.OwnedBy, "llmgopher")
		}
		if item.Created != expectedCreated[item.ID] {
			t.Fatalf("created for %q = %d, want %d", item.ID, item.Created, expectedCreated[item.ID])
		}
	}
}

func TestModelsListRoute_NoAuth_ReturnsUnauthorized(t *testing.T) {
	deps, _ := newTestDeps(nil)
	handler := api.NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestModelsListRoute_WithAuth_ReturnsListResponse(t *testing.T) {
	deps, _ := newTestDeps(nil)
	handler := api.NewRouter(deps)

	req := authedRequest(http.MethodGet, "/v1/models", "")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var body modelsListResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Object != "list" {
		t.Fatalf("object = %q, want %q", body.Object, "list")
	}
	if len(body.Data) != 0 {
		t.Fatalf("data length = %d, want 0", len(body.Data))
	}
}

func buildModelStateCache(t *testing.T) *storage.StateCache {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	providerID := uuid.New().String()
	createdAlpha := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	createdZeta := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT id, name, base_url, auth_type, has_credentials, created_at, updated_at\\s+FROM providers").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "base_url", "auth_type", "has_credentials", "created_at", "updated_at"}).
				AddRow(providerID, "openai", "https://api.openai.com/v1", "bearer", false, createdAlpha, createdAlpha),
		)

	mock.ExpectQuery("SELECT id, provider_id, name, alias, context_window, rate_limit_rps, created_at, updated_at\\s+FROM models").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "provider_id", "name", "alias", "context_window", "rate_limit_rps", "created_at", "updated_at"}).
				AddRow(uuid.NewString(), providerID, "gpt-4.1", "zeta-model", 128000, 0, createdZeta, createdZeta).
				AddRow(uuid.NewString(), providerID, "gpt-4o", "alpha-model", 128000, 0, createdAlpha, createdAlpha),
		)

	mock.ExpectQuery("SELECT id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, to_json\\(allowed_models\\), created_at, updated_at\\s+FROM api_keys\\s+WHERE is_active = TRUE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "key_hash", "name", "rate_limit_rps", "is_active", "expires_at", "metadata", "allowed_models", "created_at", "updated_at"}))

	cache := storage.NewStateCache(discardLogger)
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
