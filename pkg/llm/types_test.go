package llm

import (
	"encoding/json"
	"testing"
)

func TestChatCompletionRequest_JSONRoundTrip(t *testing.T) {
	temp := 0.7
	maxTk := 256

	original := ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "Hello"},
		},
		Temperature: &temp,
		MaxTokens:   &maxTk,
		Stream:      true,
		Stop:        json.RawMessage(`["END"]`),
		User:        "user-123",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ChatCompletionRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Model != original.Model {
		t.Errorf("model = %q, want %q", decoded.Model, original.Model)
	}
	if len(decoded.Messages) != len(original.Messages) {
		t.Fatalf("messages length = %d, want %d", len(decoded.Messages), len(original.Messages))
	}
	for i, m := range decoded.Messages {
		if m.Role != original.Messages[i].Role || m.Content != original.Messages[i].Content {
			t.Errorf("messages[%d] = %+v, want %+v", i, m, original.Messages[i])
		}
	}
	if decoded.Temperature == nil || *decoded.Temperature != temp {
		t.Errorf("temperature = %v, want %v", decoded.Temperature, temp)
	}
	if decoded.MaxTokens == nil || *decoded.MaxTokens != maxTk {
		t.Errorf("max_tokens = %v, want %v", decoded.MaxTokens, maxTk)
	}
	if !decoded.Stream {
		t.Error("stream should be true")
	}
	if decoded.User != "user-123" {
		t.Errorf("user = %q, want %q", decoded.User, "user-123")
	}
}

func TestChatCompletionRequest_MinimalPayload(t *testing.T) {
	raw := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`

	var req ChatCompletionRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if req.Model != "gpt-4o" {
		t.Errorf("model = %q, want %q", req.Model, "gpt-4o")
	}
	if len(req.Messages) != 1 {
		t.Fatalf("messages length = %d, want 1", len(req.Messages))
	}
	if req.Temperature != nil {
		t.Errorf("temperature should be nil, got %v", *req.Temperature)
	}
	if req.MaxTokens != nil {
		t.Errorf("max_tokens should be nil, got %v", *req.MaxTokens)
	}
	if req.Stream {
		t.Error("stream should default to false")
	}
}

func TestChatCompletionResponse_JSONRoundTrip(t *testing.T) {
	original := ChatCompletionResponse{
		ID:      "chatcmpl-abc123",
		Object:  "chat.completion",
		Created: 1700000000,
		Model:   "gpt-4o",
		Choices: []Choice{
			{
				Index:        0,
				Message:      &Message{Role: "assistant", Content: "Hello!"},
				FinishReason: "stop",
			},
		},
		Usage: &Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ChatCompletionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("id = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Object != "chat.completion" {
		t.Errorf("object = %q, want %q", decoded.Object, "chat.completion")
	}
	if len(decoded.Choices) != 1 {
		t.Fatalf("choices length = %d, want 1", len(decoded.Choices))
	}
	if decoded.Choices[0].Message.Content != "Hello!" {
		t.Errorf("content = %q, want %q", decoded.Choices[0].Message.Content, "Hello!")
	}
	if decoded.Usage.TotalTokens != 15 {
		t.Errorf("total_tokens = %d, want 15", decoded.Usage.TotalTokens)
	}
}

func TestChatCompletionChunk_StreamingFormat(t *testing.T) {
	chunk := ChatCompletionChunk{
		ID:      "chatcmpl-stream-1",
		Object:  "chat.completion.chunk",
		Created: 1700000000,
		Model:   "gpt-4o",
		Choices: []Choice{
			{
				Index: 0,
				Delta: &Message{Content: "Hello"},
			},
		},
	}

	data, err := json.Marshal(chunk)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ChatCompletionChunk
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Object != "chat.completion.chunk" {
		t.Errorf("object = %q, want %q", decoded.Object, "chat.completion.chunk")
	}
	if decoded.Choices[0].Delta == nil {
		t.Fatal("delta should not be nil")
	}
	if decoded.Choices[0].Delta.Content != "Hello" {
		t.Errorf("delta.content = %q, want %q", decoded.Choices[0].Delta.Content, "Hello")
	}
}

func TestAPIError_JSONFormat(t *testing.T) {
	apiErr := APIError{
		Error: APIErrorBody{
			Message: "model not found",
			Type:    "invalid_request_error",
			Code:    "model_not_found",
		},
	}

	data, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	errBody, ok := raw["error"]
	if !ok {
		t.Fatal("missing top-level 'error' key")
	}
	if errBody["message"] != "model not found" {
		t.Errorf("message = %q, want %q", errBody["message"], "model not found")
	}
	if errBody["type"] != "invalid_request_error" {
		t.Errorf("type = %q, want %q", errBody["type"], "invalid_request_error")
	}
}

func TestStop_StringAndArray(t *testing.T) {
	t.Run("string stop", func(t *testing.T) {
		raw := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stop":"END"}`
		var req ChatCompletionRequest
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		var s string
		if err := json.Unmarshal(req.Stop, &s); err != nil {
			t.Fatalf("stop as string: %v", err)
		}
		if s != "END" {
			t.Errorf("stop = %q, want %q", s, "END")
		}
	})

	t.Run("array stop", func(t *testing.T) {
		raw := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stop":["END","STOP"]}`
		var req ChatCompletionRequest
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		var arr []string
		if err := json.Unmarshal(req.Stop, &arr); err != nil {
			t.Fatalf("stop as array: %v", err)
		}
		if len(arr) != 2 || arr[0] != "END" || arr[1] != "STOP" {
			t.Errorf("stop = %v, want [END STOP]", arr)
		}
	})
}
