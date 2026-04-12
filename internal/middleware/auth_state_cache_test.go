package middleware_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/google/uuid"
)

func TestAuthWithStateCache_ExpiredKey_Returns401(t *testing.T) {
	rawKey := "sk-expired-key"
	keyHash := sha256Hex(rawKey)
	expiredAt := time.Now().Add(-1 * time.Hour).UTC()
	cache := buildAuthStateCache(t, authStateKeyRow{
		id:        "11111111-1111-1111-1111-111111111111",
		keyHash:   keyHash,
		name:      "expired-key",
		expiresAt: &expiredAt,
	})

	var reached bool
	handler := middleware.AuthWithStateCache(cache, discardLogger)(okHandler(&reached))
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if reached {
		t.Fatal("handler should not be reached for expired key")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthWithStateCache_ValidKey_Passes(t *testing.T) {
	rawKey := "sk-active-key"
	keyHash := sha256Hex(rawKey)
	expiresAt := time.Now().Add(2 * time.Hour).UTC()
	cache := buildAuthStateCache(t, authStateKeyRow{
		id:        "22222222-2222-2222-2222-222222222222",
		keyHash:   keyHash,
		name:      "active-key",
		expiresAt: &expiresAt,
	})

	var gotAPIKeyID string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotAPIKeyID = middleware.GetAPIKeyID(r.Context())
	})
	handler := middleware.AuthWithStateCache(cache, discardLogger)(inner)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotAPIKeyID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("api key id = %q, want expected id", gotAPIKeyID)
	}
}

type authStateKeyRow struct {
	id          string
	keyHash     string
	name        string
	expiresAt   *time.Time
	metadata    []byte
	models      []byte
	rateLimitPS int
}

func buildAuthStateCache(t *testing.T, key authStateKeyRow) *storage.StateCache {
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
				AddRow(uuid.NewString(), providerID, "gpt-4o", "gpt-4o", 128000, 0, now, now),
		)
	if key.metadata == nil {
		key.metadata = []byte(`{}`)
	}
	mock.ExpectQuery("SELECT id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, to_json\\(allowed_models\\), created_at, updated_at\\s+FROM api_keys\\s+WHERE is_active = TRUE").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "key_hash", "name", "rate_limit_rps", "is_active", "expires_at", "metadata", "allowed_models", "created_at", "updated_at"}).
				AddRow(key.id, key.keyHash, key.name, 100, true, key.expiresAt, key.metadata, key.models, now, now),
		)

	cache := storage.NewStateCache(discardLogger)
	ctx, cancel := context.WithCancel(context.Background())
	cache.StartPoller(ctx, db, time.Hour)
	cancel()

	if cache.Load() == nil {
		t.Fatal("cache should be primed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
	return cache
}

func sha256Hex(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}
