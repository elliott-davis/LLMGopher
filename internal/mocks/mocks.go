package mocks

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// ---------------------------------------------------------------------------
// MockProvider
// ---------------------------------------------------------------------------

// MockProvider implements llm.Provider with configurable return values.
// Set the Err field to simulate provider failures. Set ChatResponse or
// StreamBody to control successful return values.
type MockProvider struct {
	ProviderName string

	ChatResponse *llm.ChatCompletionResponse
	StreamBody   io.ReadCloser
	Err          error

	mu          sync.Mutex
	SyncCalls   []*llm.ChatCompletionRequest
	StreamCalls []*llm.ChatCompletionRequest
}

func (m *MockProvider) Name() string { return m.ProviderName }

func (m *MockProvider) ChatCompletion(_ context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	m.mu.Lock()
	m.SyncCalls = append(m.SyncCalls, req)
	m.mu.Unlock()

	if m.Err != nil {
		return nil, m.Err
	}
	return m.ChatResponse, nil
}

func (m *MockProvider) ChatCompletionStream(_ context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error) {
	m.mu.Lock()
	m.StreamCalls = append(m.StreamCalls, req)
	m.mu.Unlock()

	if m.Err != nil {
		return nil, m.Err
	}
	return m.StreamBody, nil
}

// ---------------------------------------------------------------------------
// MockRateLimiter
// ---------------------------------------------------------------------------

// MockRateLimiter implements llm.RateLimiter. Set Allowed to false to
// simulate throttling. Set Err to simulate infrastructure failures.
type MockRateLimiter struct {
	Allowed bool
	Err     error

	mu   sync.Mutex
	Keys []string
}

func (m *MockRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	m.Keys = append(m.Keys, key)
	m.mu.Unlock()

	if m.Err != nil {
		return false, m.Err
	}
	return m.Allowed, nil
}

// ---------------------------------------------------------------------------
// MockGuardrail
// ---------------------------------------------------------------------------

// MockGuardrail implements llm.Guardrail. Set Verdict to control the
// outcome. Set Err to simulate service failures (fail-closed behavior).
type MockGuardrail struct {
	Verdict *llm.GuardrailVerdict
	Err     error

	mu    sync.Mutex
	Calls []*llm.ChatCompletionRequest
}

func (m *MockGuardrail) Check(_ context.Context, req *llm.ChatCompletionRequest) (*llm.GuardrailVerdict, error) {
	m.mu.Lock()
	m.Calls = append(m.Calls, req)
	m.mu.Unlock()

	if m.Err != nil {
		return nil, m.Err
	}
	return m.Verdict, nil
}

// ---------------------------------------------------------------------------
// MockAuditLogger
// ---------------------------------------------------------------------------

// MockAuditLogger implements llm.AuditLogger and captures all logged entries.
type MockAuditLogger struct {
	Err error

	mu      sync.Mutex
	Entries []*llm.AuditEntry
}

func (m *MockAuditLogger) Log(_ context.Context, entry *llm.AuditEntry) error {
	m.mu.Lock()
	m.Entries = append(m.Entries, entry)
	m.mu.Unlock()
	return m.Err
}

func (m *MockAuditLogger) GetEntries() []*llm.AuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]*llm.AuditEntry, len(m.Entries))
	copy(cp, m.Entries)
	return cp
}

// ---------------------------------------------------------------------------
// MockBudgetTracker
// ---------------------------------------------------------------------------

// MockBudgetTracker implements llm.BudgetTracker. Set Remaining to control
// the reported budget. Set DeductErr to simulate exhausted budgets.
type MockBudgetTracker struct {
	Remaining      float64
	DeductErr      error
	GetBudgetErr   error
	MarkAlertedErr error
	BudgetState    *llm.BudgetState

	mu               sync.Mutex
	Deductions       []deduction
	MarkedAlertCalls []markedAlertCall
}

type deduction struct {
	APIKeyID string
	CostUSD  float64
}

type markedAlertCall struct {
	APIKeyID  string
	AlertedAt time.Time
}

func (m *MockBudgetTracker) RemainingBudget(_ context.Context, _ string) (float64, error) {
	return m.Remaining, nil
}

func (m *MockBudgetTracker) Deduct(_ context.Context, apiKeyID string, costUSD float64) error {
	m.mu.Lock()
	m.Deductions = append(m.Deductions, deduction{APIKeyID: apiKeyID, CostUSD: costUSD})
	m.mu.Unlock()
	return m.DeductErr
}

func (m *MockBudgetTracker) GetBudget(_ context.Context, _ string) (*llm.BudgetState, error) {
	return m.BudgetState, m.GetBudgetErr
}

func (m *MockBudgetTracker) MarkBudgetAlerted(_ context.Context, apiKeyID string, alertedAt time.Time) error {
	m.mu.Lock()
	m.MarkedAlertCalls = append(m.MarkedAlertCalls, markedAlertCall{
		APIKeyID:  apiKeyID,
		AlertedAt: alertedAt,
	})
	m.mu.Unlock()
	return m.MarkAlertedErr
}

// ---------------------------------------------------------------------------
// MockPricingLookup
// ---------------------------------------------------------------------------

// MockPricingLookup implements llm.PricingLookup with a configurable pricing
// map. If the model is not found, DefaultPricing is returned.
type MockPricingLookup struct {
	Prices         map[string]llm.ModelPricing
	DefaultPricing llm.ModelPricing
}

func NewMockPricingLookup() *MockPricingLookup {
	return &MockPricingLookup{
		Prices:         make(map[string]llm.ModelPricing),
		DefaultPricing: llm.ModelPricing{PromptPer1K: 0.01, CompletionPer1K: 0.03},
	}
}

func (m *MockPricingLookup) LookupPricing(model string) llm.ModelPricing {
	if p, ok := m.Prices[model]; ok {
		return p
	}
	return m.DefaultPricing
}
