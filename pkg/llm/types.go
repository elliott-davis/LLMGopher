package llm

import (
	"encoding/json"
	"time"
)

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
}

// Message represents a single message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
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

// Choice is a single completion choice.
type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"` // used in streaming
	FinishReason string   `json:"finish_reason,omitempty"`
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
