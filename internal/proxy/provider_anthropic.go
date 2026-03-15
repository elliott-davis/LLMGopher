package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// AnthropicProvider translates OpenAI-format requests into the Anthropic
// Messages API format and translates responses back.
type AnthropicProvider struct {
	apiKey  string
	baseURL string
	version string
	client  *http.Client
}

func NewAnthropicProvider(apiKey, baseURL string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		version: "2023-06-01",
		client:  &http.Client{},
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

// --- Anthropic wire types ---

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	StopSeqs    []string           `json:"stop_sequences,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID         string                   `json:"id"`
	Type       string                   `json:"type"`
	Role       string                   `json:"role"`
	Content    []anthropicContentBlock  `json:"content"`
	Model      string                   `json:"model"`
	StopReason string                   `json:"stop_reason"`
	Usage      anthropicUsage           `json:"usage"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// --- SSE event types for streaming ---

type anthropicStreamEvent struct {
	Type  string          `json:"type"`
	Index int             `json:"index,omitempty"`
	Delta json.RawMessage `json:"delta,omitempty"`
	Usage json.RawMessage `json:"usage,omitempty"`
}

type anthropicTextDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicStreamUsage struct {
	OutputTokens int `json:"output_tokens"`
}

type anthropicMessageStart struct {
	Type    string            `json:"type"`
	Message anthropicResponse `json:"message"`
}

// --- Translation ---

func (p *AnthropicProvider) toAnthropicRequest(req *llm.ChatCompletionRequest) *anthropicRequest {
	ar := &anthropicRequest{
		Model:       req.Model,
		MaxTokens:   4096,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	if req.MaxTokens != nil {
		ar.MaxTokens = *req.MaxTokens
	}

	if req.Stop != nil {
		var stops []string
		if err := json.Unmarshal(req.Stop, &stops); err != nil {
			var single string
			if err := json.Unmarshal(req.Stop, &single); err == nil {
				stops = []string{single}
			}
		}
		ar.StopSeqs = stops
	}

	for _, m := range req.Messages {
		if m.Role == "system" {
			ar.System = m.Content
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}
		ar.Messages = append(ar.Messages, anthropicMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	return ar
}

func (p *AnthropicProvider) fromAnthropicResponse(ar *anthropicResponse, model string) *llm.ChatCompletionResponse {
	var text string
	for _, block := range ar.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	finishReason := mapAnthropicStopReason(ar.StopReason)

	return &llm.ChatCompletionResponse{
		ID:      "chatcmpl-" + ar.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: &llm.Message{
					Role:    "assistant",
					Content: text,
				},
				FinishReason: finishReason,
			},
		},
		Usage: &llm.Usage{
			PromptTokens:     ar.Usage.InputTokens,
			CompletionTokens: ar.Usage.OutputTokens,
			TotalTokens:      ar.Usage.InputTokens + ar.Usage.OutputTokens,
		},
	}
}

func mapAnthropicStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	default:
		return reason
	}
}

// --- Sync ---

func (p *AnthropicProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	ar := p.toAnthropicRequest(req)
	ar.Stream = false

	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	p.setHeaders(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readUpstreamError(resp)
	}

	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}

	return p.fromAnthropicResponse(&result, req.Model), nil
}

// --- Streaming ---

func (p *AnthropicProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error) {
	ar := p.toAnthropicRequest(req)
	ar.Stream = true

	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	p.setHeaders(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic stream request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, readUpstreamError(resp)
	}

	// Return a translator that reads Anthropic SSE events and re-emits
	// them as OpenAI-compatible SSE events.
	pr, pw := io.Pipe()
	go p.translateStream(resp.Body, pw, req.Model)
	return pr, nil
}

// translateStream reads Anthropic SSE events from src and writes
// OpenAI-compatible SSE events to dst.
func (p *AnthropicProvider) translateStream(src io.ReadCloser, dst *io.PipeWriter, model string) {
	defer src.Close()
	defer dst.Close()

	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var msgID string
	created := time.Now().Unix()

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			fmt.Fprintf(dst, "data: [DONE]\n\n")
			return
		}

		// Try to detect the event type from the JSON.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(data), &raw); err != nil {
			continue
		}

		var eventType string
		if t, ok := raw["type"]; ok {
			json.Unmarshal(t, &eventType)
		}

		switch eventType {
		case "message_start":
			var ms anthropicMessageStart
			if err := json.Unmarshal([]byte(data), &ms); err == nil {
				msgID = ms.Message.ID
			}

		case "content_block_delta":
			var evt anthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &evt); err != nil {
				continue
			}
			var delta anthropicTextDelta
			if err := json.Unmarshal(evt.Delta, &delta); err != nil {
				continue
			}

			chunk := llm.ChatCompletionChunk{
				ID:      "chatcmpl-" + msgID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []llm.Choice{
					{
						Index: 0,
						Delta: &llm.Message{
							Content: delta.Text,
						},
					},
				},
			}
			chunkJSON, _ := json.Marshal(chunk)
			fmt.Fprintf(dst, "data: %s\n\n", chunkJSON)

		case "message_delta":
			var evt anthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &evt); err != nil {
				continue
			}

			var stopReason string
			var deltaMap map[string]interface{}
			if err := json.Unmarshal(evt.Delta, &deltaMap); err == nil {
				if sr, ok := deltaMap["stop_reason"].(string); ok {
					stopReason = mapAnthropicStopReason(sr)
				}
			}

			chunk := llm.ChatCompletionChunk{
				ID:      "chatcmpl-" + msgID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []llm.Choice{
					{
						Index:        0,
						Delta:        &llm.Message{},
						FinishReason: stopReason,
					},
				},
			}

			if evt.Usage != nil {
				var usage anthropicStreamUsage
				if err := json.Unmarshal(evt.Usage, &usage); err == nil {
					chunk.Usage = &llm.Usage{
						CompletionTokens: usage.OutputTokens,
					}
				}
			}

			chunkJSON, _ := json.Marshal(chunk)
			fmt.Fprintf(dst, "data: %s\n\n", chunkJSON)

		case "message_stop":
			fmt.Fprintf(dst, "data: [DONE]\n\n")
			return
		}
	}

	// If the scanner ends without message_stop, still send DONE.
	fmt.Fprintf(dst, "data: [DONE]\n\n")
}

func (p *AnthropicProvider) setHeaders(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", p.apiKey)
	r.Header.Set("Anthropic-Version", p.version)
}
