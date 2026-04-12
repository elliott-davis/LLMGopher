package proxy

import (
	"context"
	"log/slog"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// CostWorker handles asynchronous post-request accounting: token counting,
// cost calculation, budget deduction, and audit logging.
type CostWorker struct {
	auditLogger   llm.AuditLogger
	budgetTracker llm.BudgetTracker
	pricing       llm.PricingLookup
	tokenCounter  *TokenCounter
	logger        *slog.Logger
}

func NewCostWorker(
	auditLogger llm.AuditLogger,
	budgetTracker llm.BudgetTracker,
	pricing llm.PricingLookup,
	tokenCounter *TokenCounter,
	logger *slog.Logger,
) *CostWorker {
	return &CostWorker{
		auditLogger:   auditLogger,
		budgetTracker: budgetTracker,
		pricing:       pricing,
		tokenCounter:  tokenCounter,
		logger:        logger,
	}
}

// RequestMeta captures the context needed for post-request accounting.
type RequestMeta struct {
	RequestID    string
	APIKeyID     string
	Model        string
	ProviderName string
	Messages     []llm.Message
	Tools        []llm.Tool
	Streaming    bool
	StartTime    time.Time
}

// RecordSync is called after a non-streaming response completes.
// The response already contains usage data from the provider.
func (cw *CostWorker) RecordSync(meta *RequestMeta, resp *llm.ChatCompletionResponse, statusCode int) {
	go cw.recordSyncAsync(meta, resp, statusCode)
}

func (cw *CostWorker) recordSyncAsync(meta *RequestMeta, resp *llm.ChatCompletionResponse, statusCode int) {
	promptTokens := 0
	outputTokens := 0

	if resp.Usage != nil {
		promptTokens = resp.Usage.PromptTokens
		outputTokens = resp.Usage.CompletionTokens
	}

	// Fall back to local counting if the provider didn't report usage.
	if promptTokens == 0 {
		promptTokens = cw.tokenCounter.CountPromptTokens(meta.Model, meta.Messages, meta.Tools...)
	}
	if outputTokens == 0 && len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		outputTokens = cw.tokenCounter.CountTextTokens(meta.Model, resp.Choices[0].Message.ContentString())
	}

	totalTokens := promptTokens + outputTokens
	p := cw.pricing.LookupPricing(meta.Model)
	cost := p.CostUSD(promptTokens, outputTokens)

	cw.deductAndLog(meta, promptTokens, outputTokens, totalTokens, cost, statusCode, "")
}

// RecordStream is called after a streaming response completes.
// It blocks on the StreamInterceptor's Result() to get accumulated output,
// then performs accounting asynchronously.
func (cw *CostWorker) RecordStream(meta *RequestMeta, interceptor *StreamInterceptor, statusCode int) {
	go cw.recordStreamAsync(meta, interceptor, statusCode)
}

func (cw *CostWorker) recordStreamAsync(meta *RequestMeta, interceptor *StreamInterceptor, statusCode int) {
	// Block until the stream is fully consumed.
	result := interceptor.Result()

	promptTokens := cw.tokenCounter.CountPromptTokens(meta.Model, meta.Messages, meta.Tools...)
	outputTokens := result.OutputTokens

	totalTokens := promptTokens + outputTokens
	p := cw.pricing.LookupPricing(meta.Model)
	cost := p.CostUSD(promptTokens, outputTokens)

	cw.deductAndLog(meta, promptTokens, outputTokens, totalTokens, cost, statusCode, "")
}

// RecordError logs a failed request (no cost deduction).
func (cw *CostWorker) RecordError(meta *RequestMeta, statusCode int, errMsg string) {
	go cw.deductAndLog(meta, 0, 0, 0, 0, statusCode, errMsg)
}

func (cw *CostWorker) deductAndLog(
	meta *RequestMeta,
	promptTokens, outputTokens, totalTokens int,
	cost float64,
	statusCode int,
	errMsg string,
) {
	latency := time.Since(meta.StartTime)

	// Deduct from budget (non-blocking on failure — we still log).
	if cost > 0 && meta.APIKeyID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := cw.budgetTracker.Deduct(ctx, meta.APIKeyID, cost); err != nil {
			cw.logger.Warn("budget deduction failed",
				"error", err,
				"api_key_id", meta.APIKeyID,
				"cost_usd", cost,
				"request_id", meta.RequestID,
			)
		}
	}

	entry := &llm.AuditEntry{
		RequestID:    meta.RequestID,
		APIKeyID:     meta.APIKeyID,
		Model:        meta.Model,
		Provider:     meta.ProviderName,
		PromptTokens: promptTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		CostUSD:      cost,
		StatusCode:   statusCode,
		Latency:      latency,
		CreatedAt:    meta.StartTime,
		Streaming:    meta.Streaming,
		ErrorMessage: errMsg,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cw.auditLogger.Log(ctx, entry); err != nil {
		cw.logger.Error("failed to log audit entry",
			"error", err,
			"request_id", meta.RequestID,
		)
	}

	cw.logger.Info("request completed",
		"request_id", meta.RequestID,
		"model", meta.Model,
		"provider", meta.ProviderName,
		"prompt_tokens", promptTokens,
		"output_tokens", outputTokens,
		"cost_usd", cost,
		"latency_ms", latency.Milliseconds(),
		"streaming", meta.Streaming,
		"status", statusCode,
	)
}
