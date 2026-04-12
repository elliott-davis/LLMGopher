package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
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
	Model       string               `json:"model"`
	Messages    []anthropicMessage   `json:"messages"`
	System      string               `json:"system,omitempty"`
	MaxTokens   int                  `json:"max_tokens"`
	Temperature *float64             `json:"temperature,omitempty"`
	TopP        *float64             `json:"top_p,omitempty"`
	Stream      bool                 `json:"stream,omitempty"`
	StopSeqs    []string             `json:"stop_sequences,omitempty"`
	Tools       []anthropicTool      `json:"tools,omitempty"`
	ToolChoice  *anthropicToolChoice `json:"tool_choice,omitempty"`
}

type anthropicMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicToolChoice struct {
	Type string `json:"type"`           // "auto" | "any" | "none" | "tool"
	Name string `json:"name,omitempty"` // only when type="tool"
}

type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	StopReason string                  `json:"stop_reason"`
	Usage      anthropicUsage          `json:"usage"`
}

type anthropicContentBlock struct {
	Type      string                `json:"type"`
	Text      string                `json:"text,omitempty"`
	Source    *anthropicImageSource `json:"source,omitempty"`
	ID        string                `json:"id,omitempty"`          // tool_use block id
	Name      string                `json:"name,omitempty"`        // tool_use function name
	Input     json.RawMessage       `json:"input,omitempty"`       // tool_use arguments
	ToolUseID string                `json:"tool_use_id,omitempty"` // tool_result reference
	Content   json.RawMessage       `json:"content,omitempty"`     // tool_result payload
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	URL       string `json:"url,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
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

type anthropicInputJSONDelta struct {
	Type        string `json:"type"`
	PartialJSON string `json:"partial_json"`
}

type anthropicStreamUsage struct {
	OutputTokens int `json:"output_tokens"`
}

type anthropicMessageStart struct {
	Type    string            `json:"type"`
	Message anthropicResponse `json:"message"`
}

type anthropicContentBlockStart struct {
	Type         string                `json:"type"`
	Index        int                   `json:"index"`
	ContentBlock anthropicContentBlock `json:"content_block"`
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

	// Translate tools.
	for _, t := range req.Tools {
		schema := t.Function.Parameters
		if schema == nil {
			schema = json.RawMessage(`{}`)
		}
		ar.Tools = append(ar.Tools, anthropicTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: schema,
		})
	}

	// Translate tool_choice.
	if req.ToolChoice != nil {
		ar.ToolChoice = parseAnthropicToolChoice(req.ToolChoice)
	}

	for _, m := range req.Messages {
		if m.Role == "system" {
			ar.System = m.ContentString()
			continue
		}

		switch m.Role {
		case "tool":
			// Tool result → user message with a tool_result content block.
			payload := m.Content
			if payload == nil {
				payload = json.RawMessage(`""`)
			}
			block := anthropicContentBlock{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Content:   payload,
			}
			content, _ := json.Marshal([]anthropicContentBlock{block})
			ar.Messages = append(ar.Messages, anthropicMessage{
				Role:    "user",
				Content: content,
			})

		case "assistant":
			if len(m.ToolCalls) > 0 {
				// Mix text and tool_use content blocks.
				blocks, ok := anthropicBlocksFromMessage(m)
				if !ok {
					if cs := m.ContentString(); cs != "" {
						blocks = append(blocks, anthropicContentBlock{Type: "text", Text: cs})
					}
				}
				for _, tc := range m.ToolCalls {
					input := json.RawMessage(`{}`)
					if tc.Function.Arguments != "" {
						input = json.RawMessage(tc.Function.Arguments)
					}
					blocks = append(blocks, anthropicContentBlock{
						Type:  "tool_use",
						ID:    tc.ID,
						Name:  tc.Function.Name,
						Input: input,
					})
				}
				content, _ := json.Marshal(blocks)
				ar.Messages = append(ar.Messages, anthropicMessage{
					Role:    "assistant",
					Content: content,
				})
			} else {
				content := anthropicMessageContent(m)
				ar.Messages = append(ar.Messages, anthropicMessage{
					Role:    "assistant",
					Content: content,
				})
			}

		default: // user
			content := anthropicMessageContent(m)
			ar.Messages = append(ar.Messages, anthropicMessage{
				Role:    "user",
				Content: content,
			})
		}
	}

	return ar
}

// parseAnthropicToolChoice converts the OpenAI tool_choice field to Anthropic format.
func parseAnthropicToolChoice(raw json.RawMessage) *anthropicToolChoice {
	// String forms: "none", "auto", "required"
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		switch s {
		case "none":
			return &anthropicToolChoice{Type: "none"}
		case "auto":
			return &anthropicToolChoice{Type: "auto"}
		case "required":
			return &anthropicToolChoice{Type: "any"}
		}
		return nil
	}

	// Object form: {"type": "function", "function": {"name": "..."}}
	var obj struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil && obj.Type == "function" {
		return &anthropicToolChoice{Type: "tool", Name: obj.Function.Name}
	}
	return nil
}

func anthropicMessageContent(m llm.Message) json.RawMessage {
	if blocks, ok := anthropicBlocksFromMessage(m); ok {
		content, _ := json.Marshal(blocks)
		return content
	}
	content, _ := json.Marshal(m.ContentString())
	return json.RawMessage(content)
}

func anthropicBlocksFromMessage(m llm.Message) ([]anthropicContentBlock, bool) {
	trimmed := bytes.TrimSpace(m.Content)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, false
	}

	parts, err := m.ContentParts()
	if err != nil {
		return nil, false
	}

	blocks := make([]anthropicContentBlock, 0, len(parts))
	for _, part := range parts {
		switch part.Type {
		case "text":
			blocks = append(blocks, anthropicContentBlock{
				Type: "text",
				Text: part.Text,
			})
		case "image_url":
			if part.ImageURL == nil || part.ImageURL.URL == "" {
				continue
			}
			imageBlock, ok := anthropicImageBlockFromURL(part.ImageURL.URL)
			if !ok {
				continue
			}
			blocks = append(blocks, imageBlock)
		}
	}
	return blocks, true
}

func anthropicImageBlockFromURL(rawURL string) (anthropicContentBlock, bool) {
	if strings.HasPrefix(rawURL, "https://") {
		return anthropicContentBlock{
			Type: "image",
			Source: &anthropicImageSource{
				Type: "url",
				URL:  rawURL,
			},
		}, true
	}

	mediaType, data, ok := parseAnthropicImageDataURI(rawURL)
	if !ok {
		return anthropicContentBlock{}, false
	}

	return anthropicContentBlock{
		Type: "image",
		Source: &anthropicImageSource{
			Type:      "base64",
			MediaType: mediaType,
			Data:      data,
		},
	}, true
}

func parseAnthropicImageDataURI(uri string) (string, string, bool) {
	if !strings.HasPrefix(uri, "data:image/") {
		return "", "", false
	}

	idx := strings.Index(uri, ";base64,")
	if idx <= len("data:") {
		return "", "", false
	}

	mediaType := uri[len("data:"):idx]
	data := uri[idx+len(";base64,"):]
	if mediaType == "" || data == "" {
		return "", "", false
	}
	if _, err := base64.StdEncoding.DecodeString(data); err != nil {
		return "", "", false
	}
	return mediaType, data, true
}

func (p *AnthropicProvider) fromAnthropicResponse(ar *anthropicResponse, model string) *llm.ChatCompletionResponse {
	var text string
	var toolCalls []llm.ToolCall

	for _, block := range ar.Content {
		switch block.Type {
		case "text":
			text += block.Text
		case "tool_use":
			argsJSON, _ := json.Marshal(block.Input)
			toolCalls = append(toolCalls, llm.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: llm.FunctionCall{
					Name:      block.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	finishReason := mapAnthropicStopReason(ar.StopReason)

	msg := &llm.Message{
		Role:    "assistant",
		Content: llm.StringContent(text),
	}
	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	return &llm.ChatCompletionResponse{
		ID:      "chatcmpl-" + ar.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []llm.Choice{
			{
				Index:        0,
				Message:      msg,
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
	case "tool_use":
		return "tool_calls"
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

		// Detect event type.
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

		case "content_block_start":
			var blockStart anthropicContentBlockStart
			if err := json.Unmarshal([]byte(data), &blockStart); err != nil {
				continue
			}
			if blockStart.ContentBlock.Type != "tool_use" {
				continue
			}
			// Emit initial tool call chunk with id, type, and function name.
			chunk := llm.ChatCompletionChunk{
				ID:      "chatcmpl-" + msgID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []llm.Choice{
					{
						Index: 0,
						Delta: &llm.MessageDelta{
							ToolCalls: []llm.ToolCallDelta{
								{
									Index: blockStart.Index,
									ID:    blockStart.ContentBlock.ID,
									Type:  "function",
									Function: llm.FunctionCall{
										Name: blockStart.ContentBlock.Name,
									},
								},
							},
						},
					},
				},
			}
			chunkJSON, _ := json.Marshal(chunk)
			fmt.Fprintf(dst, "data: %s\n\n", chunkJSON)

		case "content_block_delta":
			var evt anthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &evt); err != nil {
				continue
			}

			// Determine delta type.
			var rawDelta map[string]json.RawMessage
			if err := json.Unmarshal(evt.Delta, &rawDelta); err != nil {
				continue
			}
			var deltaType string
			if t, ok := rawDelta["type"]; ok {
				json.Unmarshal(t, &deltaType)
			}

			switch deltaType {
			case "text_delta":
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
							Delta: &llm.MessageDelta{
								Content: llm.StringContent(delta.Text),
							},
						},
					},
				}
				chunkJSON, _ := json.Marshal(chunk)
				fmt.Fprintf(dst, "data: %s\n\n", chunkJSON)

			case "input_json_delta":
				var toolDelta anthropicInputJSONDelta
				if err := json.Unmarshal(evt.Delta, &toolDelta); err != nil {
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
							Delta: &llm.MessageDelta{
								ToolCalls: []llm.ToolCallDelta{
									{
										Index: evt.Index,
										Function: llm.FunctionCall{
											Arguments: toolDelta.PartialJSON,
										},
									},
								},
							},
						},
					},
				}
				chunkJSON, _ := json.Marshal(chunk)
				fmt.Fprintf(dst, "data: %s\n\n", chunkJSON)
			}

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
						Delta:        &llm.MessageDelta{},
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
