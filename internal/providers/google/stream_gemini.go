package google

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	gemini "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// streamGeminiToSSE reads from a Gemini GenerateContentResponseIterator and
// writes OpenAI-compatible SSE events to the pipe writer. It runs in its own
// goroutine and closes the writer when done.
func streamGeminiToSSE(iter *gemini.GenerateContentResponseIterator, pw *io.PipeWriter, model, chatID string) {
	defer pw.Close()

	created := time.Now().Unix()

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errChunk := llm.APIError{
				Error: llm.APIErrorBody{
					Message: err.Error(),
					Type:    "upstream_error",
				},
			}
			errJSON, _ := json.Marshal(errChunk)
			fmt.Fprintf(pw, "data: %s\n\n", errJSON)
			fmt.Fprint(pw, "data: [DONE]\n\n")
			return
		}

		for _, cand := range resp.Candidates {
			text := extractGeminiText(cand)
			finishReason := mapGeminiFinishReason(cand.FinishReason)

			chunk := llm.ChatCompletionChunk{
				ID:      chatID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []llm.Choice{
					{
						Index: int(cand.Index),
						Delta: &llm.Message{
							Content: text,
						},
						FinishReason: finishReason,
					},
				},
			}

			chunkJSON, _ := json.Marshal(chunk)
			fmt.Fprintf(pw, "data: %s\n\n", chunkJSON)
		}
	}

	// After the iterator is exhausted, extract final usage from the merged response.
	merged := iter.MergedResponse()
	if merged != nil && merged.UsageMetadata != nil {
		finalChunk := llm.ChatCompletionChunk{
			ID:      chatID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []llm.Choice{
				{
					Index:        0,
					Delta:        &llm.Message{},
					FinishReason: "stop",
				},
			},
			Usage: &llm.Usage{
				PromptTokens:     int(merged.UsageMetadata.PromptTokenCount),
				CompletionTokens: int(merged.UsageMetadata.CandidatesTokenCount),
				TotalTokens:      int(merged.UsageMetadata.TotalTokenCount),
			},
		}
		finalJSON, _ := json.Marshal(finalChunk)
		fmt.Fprintf(pw, "data: %s\n\n", finalJSON)
	}

	fmt.Fprint(pw, "data: [DONE]\n\n")
}
