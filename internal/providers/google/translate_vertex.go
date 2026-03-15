package google

import (
	"strings"

	vertex "cloud.google.com/go/vertexai/genai"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// buildVertexModel configures a Vertex AI GenerativeModel from the canonical
// request. It sets generation parameters and extracts system instructions from
// the message history, returning the model and the non-system content history.
func buildVertexModel(client *vertex.Client, modelName string, req *llm.ChatCompletionRequest) (*vertex.GenerativeModel, []*vertex.Content) {
	model := client.GenerativeModel(modelName)

	if req.Temperature != nil {
		model.SetTemperature(float32(*req.Temperature))
	}
	if req.TopP != nil {
		model.SetTopP(float32(*req.TopP))
	}
	if req.MaxTokens != nil {
		model.SetMaxOutputTokens(int32(*req.MaxTokens))
	}
	if req.N > 0 {
		model.SetCandidateCount(int32(req.N))
	}

	if req.Stop != nil {
		model.StopSequences = parseStopSequences(req.Stop)
	}

	var systemParts []vertex.Part
	var history []*vertex.Content

	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			systemParts = append(systemParts, vertex.Text(m.Content))
		case "assistant":
			history = append(history, &vertex.Content{
				Role:  "model",
				Parts: []vertex.Part{vertex.Text(m.Content)},
			})
		default:
			history = append(history, &vertex.Content{
				Role:  "user",
				Parts: []vertex.Part{vertex.Text(m.Content)},
			})
		}
	}

	if len(systemParts) > 0 {
		model.SystemInstruction = &vertex.Content{Parts: systemParts}
	}

	return model, history
}

// vertexResponseToOpenAI converts a Vertex AI GenerateContentResponse into
// an OpenAI-compatible ChatCompletionResponse.
func vertexResponseToOpenAI(resp *vertex.GenerateContentResponse, model string, id string, created int64) *llm.ChatCompletionResponse {
	out := &llm.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
	}

	for i, cand := range resp.Candidates {
		text := extractVertexText(cand)
		out.Choices = append(out.Choices, llm.Choice{
			Index: i,
			Message: &llm.Message{
				Role:    "assistant",
				Content: text,
			},
			FinishReason: mapVertexFinishReason(cand.FinishReason),
		})
	}

	if resp.UsageMetadata != nil {
		out.Usage = &llm.Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return out
}

func extractVertexText(cand *vertex.Candidate) string {
	if cand.Content == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range cand.Content.Parts {
		if t, ok := p.(vertex.Text); ok {
			b.WriteString(string(t))
		}
	}
	return b.String()
}

func mapVertexFinishReason(r vertex.FinishReason) string {
	switch r {
	case vertex.FinishReasonStop:
		return "stop"
	case vertex.FinishReasonMaxTokens:
		return "length"
	case vertex.FinishReasonSafety:
		return "content_filter"
	case vertex.FinishReasonRecitation:
		return "content_filter"
	default:
		if r == vertex.FinishReasonUnspecified {
			return ""
		}
		return "stop"
	}
}
