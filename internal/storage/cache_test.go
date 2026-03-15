package storage

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestStateCacheRefresh_SuccessSwapsState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	providerID := uuid.New()

	expectStateQueries(mock, providerID, now)

	cache := NewStateCache(discardLogger())
	if err := cache.refresh(context.Background(), db); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	state := cache.Load()
	if state == nil {
		t.Fatal("state is nil")
	}
	if len(state.Providers) != 1 {
		t.Fatalf("providers = %d, want 1", len(state.Providers))
	}
	if len(state.Models) != 1 {
		t.Fatalf("models = %d, want 1", len(state.Models))
	}
	if len(state.APIKeys) != 1 {
		t.Fatalf("api keys = %d, want 1", len(state.APIKeys))
	}

	model, ok := state.Models["production-chat"]
	if !ok {
		t.Fatal("expected alias production-chat in models map")
	}
	if model.Name != "gpt-4o" {
		t.Fatalf("resolved model name = %q, want %q", model.Name, "gpt-4o")
	}

	key, ok := state.APIKeys["hash-sk-prod"]
	if !ok {
		t.Fatal("expected hash-sk-prod in api keys map")
	}
	if key.ID != "key-001" {
		t.Fatalf("api key id = %q, want %q", key.ID, "key-001")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestStateCacheRefresh_FailureKeepsLastKnownGoodState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	providerID := uuid.New()

	expectStateQueries(mock, providerID, now)

	cache := NewStateCache(discardLogger())
	if err := cache.refresh(context.Background(), db); err != nil {
		t.Fatalf("initial refresh: %v", err)
	}
	oldState := cache.Load()
	if oldState == nil {
		t.Fatal("initial state is nil")
	}

	mock.ExpectQuery("SELECT id, name, base_url, auth_type, has_credentials, created_at, updated_at\\s+FROM providers").
		WillReturnError(errors.New("db unavailable"))

	if err := cache.refresh(context.Background(), db); err == nil {
		t.Fatal("expected refresh error, got nil")
	}

	if got := cache.Load(); got != oldState {
		t.Fatal("state pointer changed after failed refresh; expected last known good state")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestStateCacheStartPoller_CanceledContextDoesNotMutateState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	cache := NewStateCache(discardLogger())
	initial := &GatewayState{
		APIKeys:   map[string]*llm.APIKey{},
		Models:    map[string]*llm.Model{},
		Providers: map[uuid.UUID]*llm.ProviderConfig{},
	}
	cache.state.Store(initial)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cache.StartPoller(ctx, db, 5*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	if got := cache.Load(); got != initial {
		t.Fatal("state changed even though poller context was canceled")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL activity after context cancellation: %v", err)
	}
}

func expectStateQueries(mock sqlmock.Sqlmock, providerID uuid.UUID, now time.Time) {
	mock.ExpectQuery("SELECT id, name, base_url, auth_type, has_credentials, created_at, updated_at\\s+FROM providers").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "base_url", "auth_type", "has_credentials", "created_at", "updated_at"}).
				AddRow(providerID.String(), "openai", "https://api.openai.com", "bearer", false, now, now),
		)

	mock.ExpectQuery("SELECT id, provider_id, name, alias, context_window, created_at, updated_at\\s+FROM models").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "provider_id", "name", "alias", "context_window", "created_at", "updated_at"}).
				AddRow(uuid.NewString(), providerID.String(), "gpt-4o", "production-chat", 128000, now, now),
		)

	mock.ExpectQuery("SELECT id, key_hash, name, rate_limit_rps, is_active, created_at, updated_at\\s+FROM api_keys\\s+WHERE is_active = TRUE").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "key_hash", "name", "rate_limit_rps", "is_active", "created_at", "updated_at"}).
				AddRow("key-001", "hash-sk-prod", "production", 100, true, now, now),
		)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(noopWriter{}, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }
