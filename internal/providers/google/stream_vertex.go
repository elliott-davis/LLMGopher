package google

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	vertex "cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/iterator"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// streamVertexToSSE reads from a Vertex AI GenerateContentResponseIterator and
// writes OpenAI-compatible SSE events to the pipe writer. It runs in its own
// goroutine and closes the writer when done.
func streamVertexToSSE(iter *vertex.GenerateContentResponseIterator, pw *io.PipeWriter, model, chatID string) {
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
			text := extractVertexText(cand)
			toolCalls := extractVertexToolCalls(cand)
			finishReason := mapVertexFinishReason(cand.FinishReason)
			if len(toolCalls) > 0 && finishReason == "stop" {
				finishReason = "tool_calls"
			}

			delta := &llm.MessageDelta{}
			if text != "" {
				delta.Content = llm.StringContent(text)
			}
			for i, tc := range toolCalls {
				delta.ToolCalls = append(delta.ToolCalls, llm.ToolCallDelta{
					Index: i,
					ID:    tc.ID,
					Type:  "function",
					Function: llm.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}

			chunk := llm.ChatCompletionChunk{
				ID:      chatID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   model,
				Choices: []llm.Choice{
					{
						Index:        int(cand.Index),
						Delta:        delta,
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
					Delta:        &llm.MessageDelta{},
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
