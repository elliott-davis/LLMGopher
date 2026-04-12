package llm

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// --- Tool / function-calling types ---

// Tool represents a callable tool (currently only "function" type).
type Tool struct {
	Type     string   `json:"type"` // "function"
	Function Function `json:"function"`
}

// Function describes a callable function.
type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"` // JSON Schema object
}

// ToolCall is a fully-resolved tool invocation returned by the model.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall holds the name and arguments for a resolved tool call.
type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"` // JSON string
}

// ToolCallDelta is a streaming tool-call fragment.
// Index identifies which tool call in the array this fragment belongs to.
type ToolCallDelta struct {
	Index    int          `json:"index"`
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function FunctionCall `json:"function,omitempty"`
}

// StringContent creates a json.RawMessage containing the JSON-encoded form of s.
// Use this to set Message.Content from a plain Go string.
func StringContent(s string) json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}

// --- OpenAI-compatible request types (ingress) ---

// ChatCompletionRequest mirrors the OpenAI chat completions request body.
// This is the canonical internal representation used across the gateway.
type ChatCompletionRequest struct {
	Model            string          `json:"model"`
	Messages         []Message       `json:"messages"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	N                int             `json:"n,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	Stop             json.RawMessage `json:"stop,omitempty"` // string or []string
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	User             string          `json:"user,omitempty"`
	Tools            []Tool          `json:"tools,omitempty"`
	ToolChoice       json.RawMessage `json:"tool_choice,omitempty"`   // string or object
	Functions        []Function      `json:"functions,omitempty"`     // legacy
	FunctionCall     json.RawMessage `json:"function_call,omitempty"` // legacy
}

// CompletionRequest mirrors the legacy OpenAI completions request body.
// Prompt supports single string input only.
type CompletionRequest struct {
	Model            string          `json:"model"`
	Prompt           string          `json:"prompt"`
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	Stop             json.RawMessage `json:"stop,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	User             string          `json:"user,omitempty"`

	// Accepted for API compatibility; intentionally ignored by gateway routing.
	Echo     *bool  `json:"echo,omitempty"`
	Logprobs *int   `json:"logprobs,omitempty"`
	BestOf   *int   `json:"best_of,omitempty"`
	Suffix   string `json:"suffix,omitempty"`
}

// Message represents a single message in the conversation.
// Content is json.RawMessage to accommodate both plain strings ("hello") and
// structured content arrays ([{"type":"text","text":"hello"}]).
type Message struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content,omitempty"`
	Name       string          `json:"name,omitempty"`
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"` // for role="tool"
}

// ContentPart is a single item in a structured content array.
// Supported types are "text" and "image_url".
type ContentPart struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *ImageURLPart `json:"image_url,omitempty"`
}

// ImageURLPart represents an image reference for multimodal messages.
// URL may be a remote URL or a data URI.
type ImageURLPart struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// ContentParts returns the message content as structured content parts.
// If content is a plain string, a single text part is returned.
func (m *Message) ContentParts() ([]ContentPart, error) {
	if m == nil || len(m.Content) == 0 {
		return nil, nil
	}

	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return []ContentPart{
			{
				Type: "text",
				Text: s,
			},
		}, nil
	}

	var parts []ContentPart
	if err := json.Unmarshal(m.Content, &parts); err != nil {
		return nil, err
	}
	return parts, nil
}

// ContentString returns the message content as a plain string.
// If Content is a JSON string value it returns the unquoted string.
// If Content is a JSON array, it concatenates the text of all parts
// whose "type" is "text" (compatible with the vision content-part format).
// If Content is null or empty it returns "".
func (m *Message) ContentString() string {
	if m == nil || len(m.Content) == 0 {
		return ""
	}
	parts, err := m.ContentParts()
	if err != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

// MessageDelta is the streaming-delta counterpart to Message.
// Choice.Delta uses this type in SSE chunks.
type MessageDelta struct {
	Role      string          `json:"role,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	ToolCalls []ToolCallDelta `json:"tool_calls,omitempty"`
}

// ContentString returns the delta content as a plain string (same semantics as Message.ContentString).
func (d *MessageDelta) ContentString() string {
	if d == nil || len(d.Content) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(d.Content, &s); err == nil {
		return s
	}
	return concatTextParts(d.Content)
}

// concatTextParts extracts and concatenates the "text" field of all objects
// in a JSON array whose "type" is "text". Returns "" for non-array input.
func concatTextParts(raw json.RawMessage) string {
	if !rawMessageIsJSONArray(raw) {
		return ""
	}
	var parts []ContentPart
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

func rawMessageIsJSONArray(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == '['
}

// --- OpenAI-compatible response types (egress) ---

// ChatCompletionResponse mirrors the OpenAI non-streaming response.
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// CompletionChoice is a single legacy text completion choice.
type CompletionChoice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// CompletionResponse mirrors the legacy OpenAI completions response.
type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   *Usage             `json:"usage,omitempty"`
}

// Choice is a single completion choice.
type Choice struct {
	Index        int           `json:"index"`
	Message      *Message      `json:"message,omitempty"`
	Delta        *MessageDelta `json:"delta,omitempty"` // used in streaming
	FinishReason string        `json:"finish_reason,omitempty"`
}

// Usage reports token consumption.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// --- SSE streaming chunk ---

// ChatCompletionChunk is a single SSE event for streaming responses.
type ChatCompletionChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"` // present in final chunk if requested
}

// --- OpenAI-compatible embedding types ---

// EmbeddingRequest mirrors the OpenAI embeddings request body.
type EmbeddingRequest struct {
	Model          string `json:"model"`
	Input          string `json:"input"`
	EncodingFormat string `json:"encoding_format,omitempty"`
}

// EmbeddingResponse mirrors the OpenAI embeddings response body.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  *Usage          `json:"usage"`
}

// EmbeddingData holds a single embedding vector.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// --- Error response ---

// APIError is the OpenAI-compatible error envelope.
type APIError struct {
	Error APIErrorBody `json:"error"`
}

// APIErrorBody contains the error details.
type APIErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// ProviderConfig stores provider metadata and connection strategy.
type ProviderConfig struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	BaseURL        string    `json:"base_url"`
	AuthType       string    `json:"auth_type"`
	HasCredentials bool      `json:"has_credentials"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ModelConfig stores dynamic model routing metadata.
type ModelConfig struct {
	ID            string    `json:"id"`
	ProviderID    string    `json:"provider_id"`
	Name          string    `json:"name"`
	Alias         string    `json:"alias"`
	ContextWindow int       `json:"context_window"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// APIKeyConfig represents a managed API credential entry.
type APIKeyConfig struct {
	ID           string    `json:"id"`
	KeyHash      string    `json:"key_hash"`
	Name         string    `json:"name"`
	RateLimitRPS int       `json:"rate_limit_rps"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// APIKey is an alias for APIKeyConfig.
type APIKey = APIKeyConfig

// Model is an alias for ModelConfig.
type Model = ModelConfig
