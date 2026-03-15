package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
)

const defaultUpstreamURL = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"

// litellmEntry represents a single model in the LiteLLM pricing JSON.
// We only decode the fields we care about.
type litellmEntry struct {
	InputCostPerToken  *float64 `json:"input_cost_per_token"`
	OutputCostPerToken *float64 `json:"output_cost_per_token"`
	Mode               string   `json:"mode"`
	LiteLLMProvider    string   `json:"litellm_provider"`
}

// PricingSyncer periodically fetches model pricing from an upstream JSON
// source (defaults to LiteLLM's community-maintained pricing data) and
// upserts it into the database.
type PricingSyncer struct {
	store      *PgPricingStore
	httpClient *http.Client
	upstreamURL string
	logger     *slog.Logger

	stopCh chan struct{}
	done   chan struct{}
}

type PricingSyncerOption func(*PricingSyncer)

// WithUpstreamURL overrides the default LiteLLM pricing URL.
func WithUpstreamURL(url string) PricingSyncerOption {
	return func(ps *PricingSyncer) {
		ps.upstreamURL = url
	}
}

func NewPricingSyncer(store *PgPricingStore, logger *slog.Logger, opts ...PricingSyncerOption) *PricingSyncer {
	ps := &PricingSyncer{
		store: store,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		upstreamURL: defaultUpstreamURL,
		logger:      logger,
		stopCh:      make(chan struct{}),
		done:        make(chan struct{}),
	}
	for _, opt := range opts {
		opt(ps)
	}
	return ps
}

// SyncOnce fetches upstream pricing and upserts all chat models into the
// database. Returns the number of models upserted.
func (ps *PricingSyncer) SyncOnce(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ps.upstreamURL, nil)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}

	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch upstream pricing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return 0, fmt.Errorf("read response body: %w", err)
	}

	var raw map[string]litellmEntry
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, fmt.Errorf("decode upstream JSON: %w", err)
	}

	count := 0
	for model, entry := range raw {
		if !isChatModel(entry) {
			continue
		}
		if entry.InputCostPerToken == nil || entry.OutputCostPerToken == nil {
			continue
		}

		pricing := llm.ModelPricing{
			PromptPer1K:     *entry.InputCostPerToken * 1000,
			CompletionPer1K: *entry.OutputCostPerToken * 1000,
		}

		prefix := strings.ToLower(model)
		if err := ps.store.UpsertPricing(ctx, prefix, pricing, "litellm"); err != nil {
			ps.logger.Warn("failed to upsert pricing",
				"model", prefix,
				"error", err,
			)
			continue
		}
		count++
	}

	return count, nil
}

// Start launches a background loop that syncs pricing on the given interval.
// It performs an initial sync immediately.
func (ps *PricingSyncer) Start(interval time.Duration) {
	go func() {
		defer close(ps.done)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		n, err := ps.SyncOnce(ctx)
		cancel()
		if err != nil {
			ps.logger.Warn("initial pricing sync failed, will retry", "error", err)
		} else {
			ps.logger.Info("initial pricing sync complete", "models_upserted", n)
		}

		// Trigger a cache refresh after the initial sync.
		refreshCtx, refreshCancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = ps.store.refresh(refreshCtx)
		refreshCancel()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ps.stopCh:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				n, err := ps.SyncOnce(ctx)
				cancel()
				if err != nil {
					ps.logger.Warn("pricing sync failed", "error", err)
				} else {
					ps.logger.Info("pricing sync complete", "models_upserted", n)
					refreshCtx, refreshCancel := context.WithTimeout(context.Background(), 10*time.Second)
					_ = ps.store.refresh(refreshCtx)
					refreshCancel()
				}
			}
		}
	}()
}

// Close stops the background sync loop.
func (ps *PricingSyncer) Close() {
	close(ps.stopCh)
	<-ps.done
}

func isChatModel(e litellmEntry) bool {
	return e.Mode == "chat" || e.Mode == "completion" || e.Mode == ""
}
