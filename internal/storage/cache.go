package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/pkg/llm"
)

const defaultStatePollInterval = 5 * time.Second

// StateCache owns the atomically swapped gateway state snapshot.
type StateCache struct {
	state  atomic.Pointer[GatewayState]
	logger *slog.Logger
}

func NewStateCache(logger *slog.Logger) *StateCache {
	return &StateCache{
		logger: logger,
	}
}

func (c *StateCache) Load() *GatewayState {
	return c.state.Load()
}

// StartPoller refreshes gateway state at a fixed interval in a background
// goroutine. Failed refreshes are logged and ignored so the last known good
// state remains active.
func (c *StateCache) StartPoller(ctx context.Context, db *sql.DB, interval time.Duration) {
	if interval <= 0 {
		interval = defaultStatePollInterval
	}

	// Prime the cache before ticker-based polling starts.
	if err := c.refresh(ctx, db); err != nil {
		c.logger.Error("failed to prime gateway state cache", "error", err)
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("gateway state poller stopped", "reason", ctx.Err())
				return
			case <-ticker.C:
				if err := c.refresh(ctx, db); err != nil {
					c.logger.Error("gateway state refresh failed", "error", err)
					continue
				}
				if state := c.Load(); state != nil {
					log.Printf(
						"state cache poll tick: models=%d providers=%d api_keys=%d",
						len(state.Models),
						len(state.Providers),
						len(state.APIKeys),
					)
				}
			}
		}
	}()
}

func (c *StateCache) refresh(ctx context.Context, db *sql.DB) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	refreshCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	newState, err := c.loadState(refreshCtx, db)
	if err != nil {
		return err
	}

	c.state.Store(newState)
	return nil
}

func (c *StateCache) loadState(ctx context.Context, db *sql.DB) (*GatewayState, error) {
	state := &GatewayState{
		APIKeys:     make(map[string]*llm.APIKey),
		APIKeysByID: make(map[string]*llm.APIKey),
		Models:      make(map[string]*llm.Model),
		Providers:   make(map[uuid.UUID]*llm.ProviderConfig),
	}

	providers, err := db.QueryContext(ctx, `
		SELECT id, name, base_url, auth_type, has_credentials, created_at, updated_at
		FROM providers
	`)
	if err != nil {
		return nil, fmt.Errorf("query providers: %w", err)
	}
	defer providers.Close()

	for providers.Next() {
		provider := &llm.ProviderConfig{}
		if err := providers.Scan(
			&provider.ID,
			&provider.Name,
			&provider.BaseURL,
			&provider.AuthType,
			&provider.HasCredentials,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan provider row: %w", err)
		}

		id, err := uuid.Parse(provider.ID)
		if err != nil {
			return nil, fmt.Errorf("parse provider id %q: %w", provider.ID, err)
		}
		state.Providers[id] = provider
	}
	if err := providers.Err(); err != nil {
		return nil, fmt.Errorf("iterate providers: %w", err)
	}

	models, err := db.QueryContext(ctx, `
		SELECT id, provider_id, name, alias, context_window, rate_limit_rps, created_at, updated_at
		FROM models
	`)
	if err != nil {
		return nil, fmt.Errorf("query models: %w", err)
	}
	defer models.Close()

	for models.Next() {
		model := &llm.Model{}
		if err := models.Scan(
			&model.ID,
			&model.ProviderID,
			&model.Name,
			&model.Alias,
			&model.ContextWindow,
			&model.RateLimitRPS,
			&model.CreatedAt,
			&model.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan model row: %w", err)
		}
		state.Models[model.Alias] = model
	}
	if err := models.Err(); err != nil {
		return nil, fmt.Errorf("iterate models: %w", err)
	}

	apiKeys, err := db.QueryContext(ctx, `
		SELECT id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, to_json(allowed_models), created_at, updated_at
		FROM api_keys
		WHERE is_active = TRUE
	`)
	if err != nil {
		return nil, fmt.Errorf("query api keys: %w", err)
	}
	defer apiKeys.Close()

	for apiKeys.Next() {
		apiKey := &llm.APIKey{}
		var metadataRaw []byte
		var allowedModelsRaw []byte
		if err := apiKeys.Scan(
			&apiKey.ID,
			&apiKey.KeyHash,
			&apiKey.Name,
			&apiKey.RateLimitRPS,
			&apiKey.IsActive,
			&apiKey.ExpiresAt,
			&metadataRaw,
			&allowedModelsRaw,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan api key row: %w", err)
		}
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &apiKey.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal api key metadata: %w", err)
			}
		}
		if len(allowedModelsRaw) > 0 && string(allowedModelsRaw) != "null" {
			if err := json.Unmarshal(allowedModelsRaw, &apiKey.AllowedModels); err != nil {
				return nil, fmt.Errorf("unmarshal api key allowed_models: %w", err)
			}
		}
		state.APIKeys[apiKey.KeyHash] = apiKey
		state.APIKeysByID[apiKey.ID] = apiKey
	}
	if err := apiKeys.Err(); err != nil {
		return nil, fmt.Errorf("iterate api keys: %w", err)
	}

	return state, nil
}
