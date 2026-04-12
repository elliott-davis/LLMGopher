package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ed007183/llmgopher/internal/middleware"
	"github.com/ed007183/llmgopher/pkg/llm"
)

type completionRequestDecode struct {
	Model            string          `json:"model"`
	Prompt           json.RawMessage `json:"prompt"`
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	Stop             json.RawMessage `json:"stop,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	User             string          `json:"user,omitempty"`
	Echo             *bool           `json:"echo,omitempty"`
	Logprobs         *int            `json:"logprobs,omitempty"`
	BestOf           *int            `json:"best_of,omitempty"`
	Suffix           string          `json:"suffix,omitempty"`
}

// ServeCompletionsHTTP serves POST /v1/completions by translating requests and
// responses to/from the chat completion path.
func (h *Handler) ServeCompletionsHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := decodeCompletionRequest(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "model is required", "invalid_request_error")
		return
	}

	resolvedModel, expectedProviderName, err := h.resolveModel(req.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}
	req.Model = resolvedModel

	var provider llm.Provider
	if expectedProviderName != "" {
		provider, err = h.registry.ResolveProvider(expectedProviderName)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err.Error(), "server_error")
			return
		}
		if !strings.EqualFold(provider.Name(), expectedProviderName) {
			writeError(w, http.StatusServiceUnavailable, "model is configured for a different provider", "server_error")
			return
		}
	} else {
		provider, err = h.registry.Resolve(req.Model)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
			return
		}
	}
	if err := enforceAllowedModel(r.Context(), h.stateCache, req.Model); err != nil {
		if errors.Is(err, errModelNotAllowed) {
			writeError(w, http.StatusForbidden, err.Error(), "permission_error")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "authorization state unavailable", "server_error")
		return
	}

	chatReq := completionToChatRequest(req)
	meta := &RequestMeta{
		RequestID:    middleware.GetRequestID(r.Context()),
		APIKeyID:     middleware.GetAPIKeyID(r.Context()),
		Model:        chatReq.Model,
		ProviderName: provider.Name(),
		Messages:     chatReq.Messages,
		Streaming:    chatReq.Stream,
		StartTime:    time.Now(),
	}

	h.logger.Info("dispatching completions request",
		"request_id", meta.RequestID,
		"model", chatReq.Model,
		"provider", provider.Name(),
		"stream", chatReq.Stream,
	)

	if chatReq.Stream {
		h.handleCompletionStream(w, r, provider, chatReq, meta)
		return
	}
	h.handleCompletionSync(w, r, provider, chatReq, meta)
}

func decodeCompletionRequest(reader io.Reader) (*llm.CompletionRequest, error) {
	var raw completionRequestDecode
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	trimmedPrompt := bytes.TrimSpace(raw.Prompt)
	if len(trimmedPrompt) == 0 {
		return nil, fmt.Errorf("prompt is required")
	}
	if trimmedPrompt[0] == '[' {
		return nil, fmt.Errorf("prompt must be a string; prompt arrays are not supported")
	}

	var prompt string
	if err := json.Unmarshal(trimmedPrompt, &prompt); err != nil {
		return nil, fmt.Errorf("prompt must be a string")
	}

	return &llm.CompletionRequest{
		Model:            raw.Model,
		Prompt:           prompt,
		MaxTokens:        raw.MaxTokens,
		Temperature:      raw.Temperature,
		TopP:             raw.TopP,
		Stream:           raw.Stream,
		Stop:             raw.Stop,
		PresencePenalty:  raw.PresencePenalty,
		FrequencyPenalty: raw.FrequencyPenalty,
		User:             raw.User,
		Echo:             raw.Echo,
		Logprobs:         raw.Logprobs,
		BestOf:           raw.BestOf,
		Suffix:           raw.Suffix,
	}, nil
}

func completionToChatRequest(req *llm.CompletionRequest) *llm.ChatCompletionRequest {
	return &llm.ChatCompletionRequest{
		Model: req.Model,
		Messages: []llm.Message{
			{Role: "user", Content: llm.StringContent(req.Prompt)},
		},
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Stream:           req.Stream,
		Stop:             req.Stop,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
		User:             req.User,
	}
}

func completionFromChatResponse(resp *llm.ChatCompletionResponse) *llm.CompletionResponse {
	choices := make([]llm.CompletionChoice, 0, len(resp.Choices))
	for _, c := range resp.Choices {
		text := ""
		if c.Message != nil {
			text = c.Message.ContentString()
		}
		choices = append(choices, llm.CompletionChoice{
			Text:         text,
			Index:        c.Index,
			FinishReason: c.FinishReason,
		})
	}

	return &llm.CompletionResponse{
		ID:      resp.ID,
		Object:  "text_completion",
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage:   resp.Usage,
	}
}

func completionChunkFromChatChunk(chunk *llm.ChatCompletionChunk) *llm.CompletionResponse {
	out := &llm.CompletionResponse{
		ID:      chunk.ID,
		Object:  "text_completion",
		Created: chunk.Created,
		Model:   chunk.Model,
		Usage:   chunk.Usage,
	}

	out.Choices = make([]llm.CompletionChoice, 0, len(chunk.Choices))
	for _, c := range chunk.Choices {
		text := ""
		if c.Delta != nil {
			text = c.Delta.ContentString()
		}
		out.Choices = append(out.Choices, llm.CompletionChoice{
			Text:         text,
			Index:        c.Index,
			FinishReason: c.FinishReason,
		})
	}
	return out
}

func (h *Handler) handleCompletionSync(
	w http.ResponseWriter,
	r *http.Request,
	provider llm.Provider,
	chatReq *llm.ChatCompletionRequest,
	meta *RequestMeta,
) {
	resp, err := provider.ChatCompletion(r.Context(), chatReq)
	if err != nil {
		h.logger.Error("provider error",
			"error", err,
			"provider", provider.Name(),
			"request_id", meta.RequestID,
		)
		h.costWorker.RecordError(meta, http.StatusBadGateway, err.Error())
		writeError(w, http.StatusBadGateway, "upstream provider error: "+err.Error(), "upstream_error")
		return
	}

	completionResp := completionFromChatResponse(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(completionResp)

	// Fire-and-forget accounting shares the same chat-based cost worker path.
	h.costWorker.RecordSync(meta, resp, http.StatusOK)
}

func (h *Handler) handleCompletionStream(
	w http.ResponseWriter,
	r *http.Request,
	provider llm.Provider,
	chatReq *llm.ChatCompletionRequest,
	meta *RequestMeta,
) {
	stream, err := provider.ChatCompletionStream(r.Context(), chatReq)
	if err != nil {
		h.logger.Error("provider stream error",
			"error", err,
			"provider", provider.Name(),
			"request_id", meta.RequestID,
		)
		h.costWorker.RecordError(meta, http.StatusBadGateway, err.Error())
		writeError(w, http.StatusBadGateway, "upstream provider error: "+err.Error(), "upstream_error")
		return
	}

	interceptor := NewStreamInterceptor(stream, chatReq.Model, h.tokenCount)

	flusher, ok := w.(http.Flusher)
	if !ok {
		interceptor.Close()
		writeError(w, http.StatusInternalServerError, "streaming not supported", "server_error")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	scanner := bufio.NewScanner(interceptor)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
			flusher.Flush()
			continue
		}

		var chatChunk llm.ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chatChunk); err != nil {
			h.logger.Debug("unable to parse stream chunk as chat completion",
				"error", err,
				"request_id", meta.RequestID,
			)
			_, _ = w.Write([]byte(line + "\n\n"))
			flusher.Flush()
			continue
		}

		completionChunk := completionChunkFromChatChunk(&chatChunk)
		encodedChunk, err := json.Marshal(completionChunk)
		if err != nil {
			h.logger.Error("failed to marshal completions stream chunk",
				"error", err,
				"request_id", meta.RequestID,
			)
			continue
		}

		_, _ = w.Write([]byte("data: "))
		_, _ = w.Write(encodedChunk)
		_, _ = w.Write([]byte("\n\n"))
		flusher.Flush()
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		h.logger.Error("stream read error",
			"error", err,
			"request_id", meta.RequestID,
		)
	}

	h.costWorker.RecordStream(meta, interceptor, http.StatusOK)
}
