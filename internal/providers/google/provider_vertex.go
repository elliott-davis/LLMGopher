package google

import (
	"context"
	"fmt"
	"io"
	"time"

	vertex "cloud.google.com/go/vertexai/genai"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// VertexProvider implements llm.Provider for Google Cloud Vertex AI.
// Models are expected to arrive with a "vertex_ai/" prefix which is stripped
// before being passed to the SDK.
type VertexProvider struct {
	clientKey VertexClientKey
	cache     *VertexClientCache
}

func NewVertexProvider(clientKey VertexClientKey, cache *VertexClientCache) *VertexProvider {
	return &VertexProvider{
		clientKey: clientKey,
		cache:     cache,
	}
}

func (p *VertexProvider) Name() string { return "vertex_ai" }

func (p *VertexProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	client, err := p.cache.Get(ctx, p.clientKey)
	if err != nil {
		return nil, fmt.Errorf("vertex client: %w", err)
	}

	modelName := stripPrefix(req.Model, "vertex_ai/")
	model, history := buildVertexModel(client, modelName, req)

	chat := model.StartChat()
	chatHistory, lastParts := splitVertexHistory(history)
	chat.History = chatHistory

	resp, err := chat.SendMessage(ctx, lastParts...)
	if err != nil {
		return nil, fmt.Errorf("vertex generate: %w", err)
	}

	id := fmt.Sprintf("chatcmpl-vertex-%d", time.Now().UnixNano())
	return vertexResponseToOpenAI(resp, modelName, id, time.Now().Unix()), nil
}

func (p *VertexProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error) {
	client, err := p.cache.Get(ctx, p.clientKey)
	if err != nil {
		return nil, fmt.Errorf("vertex client: %w", err)
	}

	modelName := stripPrefix(req.Model, "vertex_ai/")
	model, history := buildVertexModel(client, modelName, req)

	chat := model.StartChat()
	chatHistory, lastParts := splitVertexHistory(history)
	chat.History = chatHistory

	iter := chat.SendMessageStream(ctx, lastParts...)

	chatID := fmt.Sprintf("chatcmpl-vertex-%d", time.Now().UnixNano())
	pr, pw := io.Pipe()
	go streamVertexToSSE(iter, pw, modelName, chatID)
	return pr, nil
}

// splitVertexHistory separates the full content list into prior history and
// the final user turn's parts. The Google SDK's ChatSession.SendMessage
// expects only the new user parts; prior turns go into ChatSession.History.
func splitVertexHistory(contents []*vertex.Content) ([]*vertex.Content, []vertex.Part) {
	if len(contents) == 0 {
		return nil, []vertex.Part{vertex.Text("")}
	}

	last := contents[len(contents)-1]
	history := contents[:len(contents)-1]

	parts := make([]vertex.Part, len(last.Parts))
	copy(parts, last.Parts)

	return history, parts
}
