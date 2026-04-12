package llm

import (
	"context"
	"time"
)

// AuditEntry captures the metadata for a single LLM request lifecycle.
type AuditEntry struct {
	ID           int64
	RequestID    string
	APIKeyID     string
	Model        string
	Provider     string
	PromptTokens int
	OutputTokens int
	TotalTokens  int
	CostUSD      float64
	StatusCode   int
	Latency      time.Duration
	CreatedAt    time.Time
	Streaming    bool
	ErrorMessage string
}

// AuditLogger persists request audit trails for cost tracking and observability.
type AuditLogger interface {
	// Log writes an audit entry. Implementations should be non-blocking
	// (e.g. channel-buffered) so they don't add latency to the response path.
	Log(ctx context.Context, entry *AuditEntry) error
}

// BudgetState captures persisted budget configuration and current spend.
type BudgetState struct {
	APIKeyID          string
	BudgetUSD         float64
	SpentUSD          float64
	AlertThresholdPct int
	BudgetDuration    string
	BudgetResetAt     *time.Time
	LastAlertedAt     *time.Time
}

// BudgetTracker manages per-key or per-org spend limits.
type BudgetTracker interface {
	// RemainingBudget returns the remaining budget in USD for the given key.
	RemainingBudget(ctx context.Context, apiKeyID string) (float64, error)

	// Deduct subtracts costUSD from the budget. Returns an error if the
	// deduction would exceed the limit (budget exhausted).
	Deduct(ctx context.Context, apiKeyID string, costUSD float64) error

	// GetBudget returns the persisted budget state for the key.
	// A missing budget should return (nil, nil).
	GetBudget(ctx context.Context, apiKeyID string) (*BudgetState, error)

	// MarkBudgetAlerted updates the last alerted timestamp for threshold alerts.
	MarkBudgetAlerted(ctx context.Context, apiKeyID string, alertedAt time.Time) error
}

// ModelPricing holds per-token costs in USD.
type ModelPricing struct {
	PromptPer1K     float64
	CompletionPer1K float64
}

// CostUSD computes the total cost for the given token counts.
func (p ModelPricing) CostUSD(promptTokens, completionTokens int) float64 {
	return (float64(promptTokens)/1000)*p.PromptPer1K +
		(float64(completionTokens)/1000)*p.CompletionPer1K
}

// PricingLookup resolves model names to their pricing. Implementations may
// back this with a database, config file, or hardcoded table.
type PricingLookup interface {
	// LookupPricing returns the pricing for the given model. If no specific
	// match is found, a sensible default is returned.
	LookupPricing(model string) ModelPricing
}
