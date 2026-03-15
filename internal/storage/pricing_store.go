package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
)

var defaultPricing = llm.ModelPricing{PromptPer1K: 0.0100, CompletionPer1K: 0.0300}

// PgPricingStore implements llm.PricingLookup backed by PostgreSQL with an
// in-memory cache that refreshes periodically.
type PgPricingStore struct {
	db     *sql.DB
	logger *slog.Logger

	mu    sync.RWMutex
	cache map[string]llm.ModelPricing

	stopCh chan struct{}
	done   chan struct{}
}

func NewPgPricingStore(db *sql.DB, logger *slog.Logger) *PgPricingStore {
	return &PgPricingStore{
		db:     db,
		logger: logger,
		cache:  make(map[string]llm.ModelPricing),
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
}

// Load performs the initial cache population from the database. Call once at
// startup before serving traffic.
func (ps *PgPricingStore) Load(ctx context.Context) error {
	return ps.refresh(ctx)
}

// StartRefresh launches a background goroutine that re-reads pricing from the
// database on the given interval.
func (ps *PgPricingStore) StartRefresh(interval time.Duration) {
	go func() {
		defer close(ps.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ps.stopCh:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := ps.refresh(ctx); err != nil {
					ps.logger.Warn("pricing cache refresh failed", "error", err)
				}
				cancel()
			}
		}
	}()
}

// Close stops the background refresh goroutine.
func (ps *PgPricingStore) Close() {
	close(ps.stopCh)
	<-ps.done
}

func (ps *PgPricingStore) refresh(ctx context.Context) error {
	rows, err := ps.db.QueryContext(ctx,
		`SELECT model_prefix, prompt_per_1k, completion_per_1k FROM model_pricing`,
	)
	if err != nil {
		return fmt.Errorf("query model_pricing: %w", err)
	}
	defer rows.Close()

	newCache := make(map[string]llm.ModelPricing)
	for rows.Next() {
		var prefix string
		var p llm.ModelPricing
		if err := rows.Scan(&prefix, &p.PromptPer1K, &p.CompletionPer1K); err != nil {
			return fmt.Errorf("scan model_pricing row: %w", err)
		}
		newCache[prefix] = p
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate model_pricing rows: %w", err)
	}

	ps.mu.Lock()
	ps.cache = newCache
	ps.mu.Unlock()

	ps.logger.Debug("pricing cache refreshed", "models", len(newCache))
	return nil
}

// LookupPricing finds the best matching pricing for a model name.
// It tries exact match first, then longest prefix match, then falls back to a
// sensible default.
func (ps *PgPricingStore) LookupPricing(model string) llm.ModelPricing {
	m := strings.ToLower(model)

	ps.mu.RLock()
	cache := ps.cache
	ps.mu.RUnlock()

	if p, ok := cache[m]; ok {
		return p
	}

	bestKey := ""
	for key := range cache {
		if strings.HasPrefix(m, key) && len(key) > len(bestKey) {
			bestKey = key
		}
	}
	if bestKey != "" {
		return cache[bestKey]
	}

	return defaultPricing
}

// UpsertPricing inserts or updates a single model's pricing in the database.
// The in-memory cache is updated on the next refresh cycle.
func (ps *PgPricingStore) UpsertPricing(ctx context.Context, prefix string, p llm.ModelPricing, source string) error {
	_, err := ps.db.ExecContext(ctx, `
		INSERT INTO model_pricing (model_prefix, prompt_per_1k, completion_per_1k, source, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (model_prefix) DO UPDATE SET
			prompt_per_1k     = EXCLUDED.prompt_per_1k,
			completion_per_1k = EXCLUDED.completion_per_1k,
			source            = EXCLUDED.source,
			updated_at        = NOW()
	`, prefix, p.PromptPer1K, p.CompletionPer1K, source)
	return err
}
