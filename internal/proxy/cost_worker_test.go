package proxy

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/ed007183/llmgopher/internal/mocks"
	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestCostWorker_EmitsBudgetThresholdAlert(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	budgetTracker := &mocks.MockBudgetTracker{
		Remaining: 1e9,
		BudgetState: &llm.BudgetState{
			APIKeyID:          "key-1",
			BudgetUSD:         100,
			SpentUSD:          80,
			AlertThresholdPct: 80,
		},
	}

	worker := NewCostWorker(
		&mocks.MockAuditLogger{},
		budgetTracker,
		mocks.NewMockPricingLookup(),
		nil,
		logger,
	)
	meta := &RequestMeta{
		RequestID:    "req-1",
		APIKeyID:     "key-1",
		Model:        "gpt-4o-mini",
		ProviderName: "openai",
		StartTime:    time.Now().Add(-150 * time.Millisecond),
	}

	worker.deductAndLog(meta, 10, 5, 15, 0.25, 200, "")

	if len(budgetTracker.MarkedAlertCalls) != 1 {
		t.Fatalf("marked alert calls = %d, want 1", len(budgetTracker.MarkedAlertCalls))
	}
	if !strings.Contains(logBuf.String(), "budget_threshold_reached") {
		t.Fatalf("expected budget threshold warning in logs, got: %s", logBuf.String())
	}
}

func TestCostWorker_DoesNotRepeatBudgetThresholdAlertWithinCooldown(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	lastAlertedAt := time.Now().Add(-30 * time.Minute)
	budgetTracker := &mocks.MockBudgetTracker{
		Remaining: 1e9,
		BudgetState: &llm.BudgetState{
			APIKeyID:          "key-1",
			BudgetUSD:         100,
			SpentUSD:          90,
			AlertThresholdPct: 80,
			LastAlertedAt:     &lastAlertedAt,
		},
	}

	worker := NewCostWorker(
		&mocks.MockAuditLogger{},
		budgetTracker,
		mocks.NewMockPricingLookup(),
		nil,
		logger,
	)
	meta := &RequestMeta{
		RequestID:    "req-2",
		APIKeyID:     "key-1",
		Model:        "gpt-4o-mini",
		ProviderName: "openai",
		StartTime:    time.Now().Add(-150 * time.Millisecond),
	}

	worker.deductAndLog(meta, 10, 5, 15, 0.25, 200, "")

	if len(budgetTracker.MarkedAlertCalls) != 0 {
		t.Fatalf("marked alert calls = %d, want 0", len(budgetTracker.MarkedAlertCalls))
	}
	if strings.Contains(logBuf.String(), "budget_threshold_reached") {
		t.Fatalf("did not expect budget threshold warning in logs, got: %s", logBuf.String())
	}
}
