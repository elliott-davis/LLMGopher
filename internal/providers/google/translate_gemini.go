package google

import (
	"encoding/json"
	"strings"

	gemini "github.com/google/generative-ai-go/genai"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// buildGeminiModel configures a GenerativeModel from the canonical request.
// It sets generation parameters and extracts system instructions from the
// message history, returning the model and the non-system content history.
func buildGeminiModel(client *gemini.Client, modelName string, req *llm.ChatCompletionRequest) (*gemini.GenerativeModel, []*gemini.Content) {
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

	var systemParts []gemini.Part
	var history []*gemini.Content

	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			systemParts = append(systemParts, gemini.Text(m.Content))
		case "assistant":
			history = append(history, &gemini.Content{
				Role:  "model",
				Parts: []gemini.Part{gemini.Text(m.Content)},
			})
		default:
			history = append(history, &gemini.Content{
				Role:  "user",
				Parts: []gemini.Part{gemini.Text(m.Content)},
			})
		}
	}

	if len(systemParts) > 0 {
		model.SystemInstruction = &gemini.Content{Parts: systemParts}
	}

	return model, history
}

// geminiResponseToOpenAI converts a Gemini GenerateContentResponse into
// an OpenAI-compatible ChatCompletionResponse.
func geminiResponseToOpenAI(resp *gemini.GenerateContentResponse, model string, id string, created int64) *llm.ChatCompletionResponse {
	out := &llm.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
	}

	for i, cand := range resp.Candidates {
		text := extractGeminiText(cand)
		out.Choices = append(out.Choices, llm.Choice{
			Index: i,
			Message: &llm.Message{
				Role:    "assistant",
				Content: text,
			},
			FinishReason: mapGeminiFinishReason(cand.FinishReason),
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

func extractGeminiText(cand *gemini.Candidate) string {
	if cand.Content == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range cand.Content.Parts {
		if t, ok := p.(gemini.Text); ok {
			b.WriteString(string(t))
		}
	}
	return b.String()
}

func mapGeminiFinishReason(r gemini.FinishReason) string {
	switch r {
	case gemini.FinishReasonStop:
		return "stop"
	case gemini.FinishReasonMaxTokens:
		return "length"
	case gemini.FinishReasonSafety:
		return "content_filter"
	case gemini.FinishReasonRecitation:
		return "content_filter"
	default:
		if r == gemini.FinishReasonUnspecified {
			return ""
		}
		return "stop"
	}
}

func parseStopSequences(raw json.RawMessage) []string {
	var stops []string
	if err := json.Unmarshal(raw, &stops); err != nil {
		var single string
		if err := json.Unmarshal(raw, &single); err == nil {
			stops = []string{single}
		}
	}
	return stops
}
