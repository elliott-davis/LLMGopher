package llm

import (
	"context"
	"io"
)

// Provider abstracts a backend LLM service (OpenAI, Anthropic, Gemini, etc.).
// Implementations translate canonical requests into provider-specific wire formats.
type Provider interface {
	// Name returns the unique identifier for this provider (e.g. "openai", "anthropic").
	Name() string

	// ChatCompletion sends a non-streaming request and returns the full response.
	ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)

	// ChatCompletionStream sends a streaming request and returns a reader that
	// yields SSE-encoded chunks. The caller is responsible for closing the reader.
	ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (io.ReadCloser, error)
}

// EmbeddingProvider is an optional interface that a Provider may implement
// to support the /v1/embeddings endpoint.
type EmbeddingProvider interface {
	EmbedContent(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
}

// ProviderRegistry maps model names (or prefixes) to their backing Provider.
type ProviderRegistry interface {
	// Resolve returns the Provider responsible for the given model identifier.
	Resolve(model string) (Provider, error)

	// ResolveProvider returns a Provider by its canonical provider name.
	ResolveProvider(name string) (Provider, error)

	// Register adds a provider with the model patterns it serves.
	Register(provider Provider, models ...string)
}
