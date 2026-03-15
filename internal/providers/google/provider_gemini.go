package google

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	gemini "github.com/google/generative-ai-go/genai"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// GeminiProvider implements llm.Provider for the Google AI Studio (Gemini) API.
// Models are expected to arrive with a "gemini/" prefix which is stripped
// before being passed to the SDK.
type GeminiProvider struct {
	apiKey string
	cache  *GeminiClientCache
}

func NewGeminiProvider(apiKey string, cache *GeminiClientCache) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		cache:  cache,
	}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	client, err := p.cache.Get(ctx, p.apiKey)
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	modelName := stripPrefix(req.Model, "gemini/")
	model, history := buildGeminiModel(client, modelName, req)

	chat := model.StartChat()
	chatHistory, lastParts := splitGeminiHistory(history)
	chat.History = chatHistory

	resp, err := chat.SendMessage(ctx, lastParts...)
	if err != nil {
		return nil, fmt.Errorf("gemini generate: %w", err)
	}

	id := fmt.Sprintf("chatcmpl-gemini-%d", time.Now().UnixNano())
	return geminiResponseToOpenAI(resp, modelName, id, time.Now().Unix()), nil
}

func (p *GeminiProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error) {
	client, err := p.cache.Get(ctx, p.apiKey)
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	modelName := stripPrefix(req.Model, "gemini/")
	model, history := buildGeminiModel(client, modelName, req)

	chat := model.StartChat()
	chatHistory, lastParts := splitGeminiHistory(history)
	chat.History = chatHistory

	iter := chat.SendMessageStream(ctx, lastParts...)

	chatID := fmt.Sprintf("chatcmpl-gemini-%d", time.Now().UnixNano())
	pr, pw := io.Pipe()
	go streamGeminiToSSE(iter, pw, modelName, chatID)
	return pr, nil
}

// splitGeminiHistory separates the full content list into prior history and
// the final user turn's parts. The Google SDK's ChatSession.SendMessage
// expects only the new user parts; prior turns go into ChatSession.History.
func splitGeminiHistory(contents []*gemini.Content) ([]*gemini.Content, []gemini.Part) {
	if len(contents) == 0 {
		return nil, []gemini.Part{gemini.Text("")}
	}

	last := contents[len(contents)-1]
	history := contents[:len(contents)-1]

	parts := make([]gemini.Part, len(last.Parts))
	copy(parts, last.Parts)

	return history, parts
}

// EmbedContent implements llm.EmbeddingProvider for the Gemini API.
func (p *GeminiProvider) EmbedContent(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	client, err := p.cache.Get(ctx, p.apiKey)
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	modelName := stripPrefix(req.Model, "gemini/")
	em := client.EmbeddingModel(modelName)

	resp, err := em.EmbedContent(ctx, gemini.Text(req.Input))
	if err != nil {
		return nil, fmt.Errorf("gemini embed: %w", err)
	}

	if resp.Embedding == nil {
		return nil, fmt.Errorf("gemini embed: empty embedding response")
	}

	return &llm.EmbeddingResponse{
		Object: "list",
		Data: []llm.EmbeddingData{
			{
				Object:    "embedding",
				Embedding: resp.Embedding.Values,
				Index:     0,
			},
		},
		Model: modelName,
		Usage: &llm.Usage{
			PromptTokens: len(req.Input) / 4, // approximate; Gemini doesn't report embedding token usage
			TotalTokens:  len(req.Input) / 4,
		},
	}, nil
}

// stripPrefix removes the routing prefix (e.g. "gemini/" or "vertex_ai/")
// from the model name before passing it to the SDK.
func stripPrefix(model, prefix string) string {
	return strings.TrimPrefix(model, prefix)
}
