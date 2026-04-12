package middleware

import (
	"bytes"
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

// NemoGuardrail calls an external NeMo Guardrails-compatible service to
// validate prompts before they reach the LLM. The service is expected to
// return a JSON response with an "allowed" boolean and optional "reason".
//
// If the endpoint is empty, the guardrail is effectively a no-op (always allows).
type NemoGuardrail struct {
	endpoint string
	timeout  time.Duration
	client   *http.Client
	logger   *slog.Logger
}

func NewNemoGuardrail(endpoint string, timeout time.Duration, logger *slog.Logger) *NemoGuardrail {
	return &NemoGuardrail{
		endpoint: endpoint,
		timeout:  timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// nemoRequest is the payload sent to the guardrail service.
type nemoRequest struct {
	Messages []nemoMessage `json:"messages"`
	Model    string        `json:"model"`
}

type nemoMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// nemoResponse is the expected response from the guardrail service.
type nemoResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

func (ng *NemoGuardrail) Check(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.GuardrailVerdict, error) {
	if ng.endpoint == "" {
		return &llm.GuardrailVerdict{Allowed: true}, nil
	}

	checkCtx, cancel := context.WithTimeout(ctx, ng.timeout)
	defer cancel()

	msgs := make([]nemoMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = nemoMessage{Role: m.Role, Content: m.ContentString()}
	}

	payload, err := json.Marshal(nemoRequest{
		Messages: msgs,
		Model:    req.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal guardrail request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(checkCtx, http.MethodPost, ng.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create guardrail request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ng.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("guardrail service call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read guardrail response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("guardrail service returned %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var nemoResp nemoResponse
	if err := json.Unmarshal(body, &nemoResp); err != nil {
		return nil, fmt.Errorf("decode guardrail response: %w", err)
	}

	ng.logger.Debug("guardrail check completed",
		"allowed", nemoResp.Allowed,
		"reason", nemoResp.Reason,
	)

	return &llm.GuardrailVerdict{
		Allowed: nemoResp.Allowed,
		Reason:  nemoResp.Reason,
	}, nil
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
