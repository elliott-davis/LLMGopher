package internal

import (
	"context"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// NoopRateLimiter always allows requests. Used when Redis is not configured.
type NoopRateLimiter struct{}

func (NoopRateLimiter) Allow(_ context.Context, _ string) (bool, error) { return true, nil }

// NoopGuardrail always permits requests. Used when no guardrail endpoint is configured.
type NoopGuardrail struct{}

func (NoopGuardrail) Check(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.GuardrailVerdict, error) {
	return &llm.GuardrailVerdict{Allowed: true}, nil
}

// NoopAuditLogger discards audit entries. Used as a fallback.
type NoopAuditLogger struct{}

func (NoopAuditLogger) Log(_ context.Context, _ *llm.AuditEntry) error { return nil }

// NoopBudgetTracker reports unlimited budget. Used as a fallback.
type NoopBudgetTracker struct{}

func (NoopBudgetTracker) RemainingBudget(_ context.Context, _ string) (float64, error) {
	return 1e9, nil
}

func (NoopBudgetTracker) Deduct(_ context.Context, _ string, _ float64) error { return nil }

// NoopPricingLookup returns a conservative default for all models.
type NoopPricingLookup struct{}

func (NoopPricingLookup) LookupPricing(_ string) llm.ModelPricing {
	return llm.ModelPricing{PromptPer1K: 0.0100, CompletionPer1K: 0.0300}
}
