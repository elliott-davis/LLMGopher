package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// GuardrailCheck runs the prompt through the Guardrail service before
// forwarding to the LLM provider. If the guardrail rejects the prompt,
// a 400 is returned immediately.
func GuardrailCheck(guard llm.Guardrail, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if guard == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Buffer the body so we can read it for the guardrail check
			// and then replay it for the actual handler.
			body, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				http.Error(w,
					`{"error":{"message":"failed to read request body","type":"invalid_request_error"}}`,
					http.StatusBadRequest,
				)
				return
			}

			req, err := guardrailRequestFromBody(body)
			if err != nil {
				http.Error(w,
					`{"error":{"message":"invalid JSON in request body","type":"invalid_request_error"}}`,
					http.StatusBadRequest,
				)
				return
			}

			verdict, err := guard.Check(r.Context(), req)
			if err != nil {
				logger.Error("guardrail check failed",
					"error", err,
					"request_id", GetRequestID(r.Context()),
				)
				// Fail closed: reject on guardrail errors.
				http.Error(w,
					`{"error":{"message":"guardrail service unavailable","type":"server_error"}}`,
					http.StatusServiceUnavailable,
				)
				return
			}

			if !verdict.Allowed {
				logger.Info("guardrail blocked request",
					"reason", verdict.Reason,
					"request_id", GetRequestID(r.Context()),
				)
				http.Error(w,
					`{"error":{"message":"request blocked by content policy: `+verdict.Reason+`","type":"content_policy_error"}}`,
					http.StatusBadRequest,
				)
				return
			}

			// Restore the body for downstream handlers.
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}

func guardrailRequestFromBody(body []byte) (*llm.ChatCompletionRequest, error) {
	var chatReq llm.ChatCompletionRequest
	if err := json.Unmarshal(body, &chatReq); err != nil {
		return nil, err
	}
	if len(chatReq.Messages) > 0 {
		return &chatReq, nil
	}

	// /v1/completions carries prompt instead of messages. Translate it to a
	// single user message so guardrails evaluate the same user-provided text.
	var completionReq struct {
		Model  string          `json:"model"`
		Prompt json.RawMessage `json:"prompt"`
	}
	if err := json.Unmarshal(body, &completionReq); err != nil {
		return nil, err
	}
	trimmedPrompt := bytes.TrimSpace(completionReq.Prompt)
	if len(trimmedPrompt) == 0 {
		return &chatReq, nil
	}

	var prompt string
	if err := json.Unmarshal(trimmedPrompt, &prompt); err != nil {
		// Let downstream handlers return endpoint-specific validation errors.
		return &chatReq, nil
	}

	if completionReq.Prompt != nil {
		translated := &llm.ChatCompletionRequest{
			Model: completionReq.Model,
			Messages: []llm.Message{
				{Role: "user", Content: llm.StringContent(prompt)},
			},
		}
		return translated, nil
	}

	return &chatReq, nil
}
