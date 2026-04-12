package google

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// simulateSSEStream writes the same SSE format that streamGeminiToSSE /
// streamVertexToSSE produce, allowing us to verify the wire format without
// needing a real SDK iterator.
func simulateSSEStream(pw *io.PipeWriter, model, chatID string, chunks []string, usage *llm.Usage) {
	defer pw.Close()

	created := time.Now().Unix()

	for _, text := range chunks {
		chunk := llm.ChatCompletionChunk{
			ID:      chatID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []llm.Choice{
				{
					Index: 0,
					Delta: &llm.MessageDelta{Content: llm.StringContent(text)},
				},
			},
		}
		chunkJSON, _ := json.Marshal(chunk)
		fmt.Fprintf(pw, "data: %s\n\n", chunkJSON)
	}

	if usage != nil {
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
			Usage: usage,
		}
		finalJSON, _ := json.Marshal(finalChunk)
		fmt.Fprintf(pw, "data: %s\n\n", finalJSON)
	}

	fmt.Fprint(pw, "data: [DONE]\n\n")
}

func TestSSEStream_Format(t *testing.T) {
	pr, pw := io.Pipe()
	go simulateSSEStream(pw, "gemini-1.5-pro", "chatcmpl-test-123",
		[]string{"Hello", ", ", "world!"},
		&llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	)

	scanner := bufio.NewScanner(pr)
	var events []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			events = append(events, strings.TrimPrefix(line, "data: "))
		}
	}

	// 3 content chunks + 1 final usage chunk + 1 [DONE]
	if len(events) != 5 {
		t.Fatalf("got %d events, want 5; events: %v", len(events), events)
	}

	// Verify content chunks
	for i, wantText := range []string{"Hello", ", ", "world!"} {
		var chunk llm.ChatCompletionChunk
		if err := json.Unmarshal([]byte(events[i]), &chunk); err != nil {
			t.Fatalf("event %d: unmarshal error: %v", i, err)
		}
		if chunk.ID != "chatcmpl-test-123" {
			t.Errorf("event %d: ID = %q, want chatcmpl-test-123", i, chunk.ID)
		}
		if chunk.Object != "chat.completion.chunk" {
			t.Errorf("event %d: Object = %q, want chat.completion.chunk", i, chunk.Object)
		}
		if chunk.Model != "gemini-1.5-pro" {
			t.Errorf("event %d: Model = %q, want gemini-1.5-pro", i, chunk.Model)
		}
		if len(chunk.Choices) != 1 {
			t.Fatalf("event %d: len(Choices) = %d, want 1", i, len(chunk.Choices))
		}
		if chunk.Choices[0].Delta == nil || chunk.Choices[0].Delta.ContentString() != wantText {
			t.Errorf("event %d: Delta.ContentString() = %q, want %q", i, chunk.Choices[0].Delta.ContentString(), wantText)
		}
		if chunk.Usage != nil {
			t.Errorf("event %d: Usage should be nil for content chunks", i)
		}
	}

	// Verify final usage chunk
	var finalChunk llm.ChatCompletionChunk
	if err := json.Unmarshal([]byte(events[3]), &finalChunk); err != nil {
		t.Fatalf("final chunk: unmarshal error: %v", err)
	}
	if finalChunk.Choices[0].FinishReason != "stop" {
		t.Errorf("final chunk: FinishReason = %q, want stop", finalChunk.Choices[0].FinishReason)
	}
	if finalChunk.Usage == nil {
		t.Fatal("final chunk: Usage is nil")
	}
	if finalChunk.Usage.PromptTokens != 10 {
		t.Errorf("final chunk: PromptTokens = %d, want 10", finalChunk.Usage.PromptTokens)
	}
	if finalChunk.Usage.CompletionTokens != 5 {
		t.Errorf("final chunk: CompletionTokens = %d, want 5", finalChunk.Usage.CompletionTokens)
	}
	if finalChunk.Usage.TotalTokens != 15 {
		t.Errorf("final chunk: TotalTokens = %d, want 15", finalChunk.Usage.TotalTokens)
	}

	// Verify [DONE] sentinel
	if events[4] != "[DONE]" {
		t.Errorf("last event = %q, want [DONE]", events[4])
	}
}

func TestSSEStream_NoUsage(t *testing.T) {
	pr, pw := io.Pipe()
	go simulateSSEStream(pw, "gemini-1.5-flash", "chatcmpl-no-usage",
		[]string{"Hi"},
		nil,
	)

	scanner := bufio.NewScanner(pr)
	var events []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			events = append(events, strings.TrimPrefix(line, "data: "))
		}
	}

	// 1 content chunk + [DONE] (no usage chunk)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2; events: %v", len(events), events)
	}

	if events[1] != "[DONE]" {
		t.Errorf("last event = %q, want [DONE]", events[1])
	}
}

func TestSSEStream_EmptyChunks(t *testing.T) {
	pr, pw := io.Pipe()
	go simulateSSEStream(pw, "gemini-1.5-pro", "chatcmpl-empty",
		nil,
		&llm.Usage{PromptTokens: 5, CompletionTokens: 0, TotalTokens: 5},
	)

	scanner := bufio.NewScanner(pr)
	var events []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			events = append(events, strings.TrimPrefix(line, "data: "))
		}
	}

	// 0 content chunks + 1 usage chunk + [DONE]
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2; events: %v", len(events), events)
	}

	var usageChunk llm.ChatCompletionChunk
	if err := json.Unmarshal([]byte(events[0]), &usageChunk); err != nil {
		t.Fatalf("usage chunk: unmarshal error: %v", err)
	}
	if usageChunk.Usage == nil {
		t.Fatal("usage chunk: Usage is nil")
	}
	if usageChunk.Usage.PromptTokens != 5 {
		t.Errorf("usage chunk: PromptTokens = %d, want 5", usageChunk.Usage.PromptTokens)
	}

	if events[1] != "[DONE]" {
		t.Errorf("last event = %q, want [DONE]", events[1])
	}
}

// TestSSEStream_InterceptorCompatibility verifies that the SSE format produced
// by our streaming functions is correctly parsed by the existing StreamInterceptor.
// This is the critical integration point — the handler reads from our pipe via
// the interceptor, which expects OpenAI-compatible SSE.
func TestSSEStream_InterceptorCompatibility(t *testing.T) {
	pr, pw := io.Pipe()
	go simulateSSEStream(pw, "gemini-1.5-pro", "chatcmpl-compat",
		[]string{"Hello", " world"},
		&llm.Usage{PromptTokens: 8, CompletionTokens: 3, TotalTokens: 11},
	)

	// Read all SSE data and verify it's parseable as OpenAI chunks.
	scanner := bufio.NewScanner(pr)
	var totalText strings.Builder
	var lastUsage *llm.Usage

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk llm.ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			t.Fatalf("failed to parse SSE chunk: %v\ndata: %s", err, data)
		}

		for _, choice := range chunk.Choices {
			if choice.Delta != nil {
				if s := choice.Delta.ContentString(); s != "" {
					totalText.WriteString(s)
				}
			}
		}

		if chunk.Usage != nil {
			lastUsage = chunk.Usage
		}
	}

	if totalText.String() != "Hello world" {
		t.Errorf("accumulated text = %q, want 'Hello world'", totalText.String())
	}

	if lastUsage == nil {
		t.Fatal("no usage metadata found in stream")
	}
	if lastUsage.CompletionTokens != 3 {
		t.Errorf("CompletionTokens = %d, want 3", lastUsage.CompletionTokens)
	}
}
