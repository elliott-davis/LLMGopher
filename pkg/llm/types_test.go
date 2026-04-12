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
			{Role: "system", Content: StringContent("You are helpful.")},
			{Role: "user", Content: StringContent("Hello")},
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
		if m.Role != original.Messages[i].Role || m.ContentString() != original.Messages[i].ContentString() {
			t.Errorf("messages[%d] role=%q content=%q, want role=%q content=%q",
				i, m.Role, m.ContentString(), original.Messages[i].Role, original.Messages[i].ContentString())
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
	if req.Messages[0].ContentString() != "hi" {
		t.Errorf("content = %q, want hi", req.Messages[0].ContentString())
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

func TestCompletionRequest_JSONRoundTrip(t *testing.T) {
	temp := 0.3
	maxTk := 64
	echo := true
	logprobs := 2
	bestOf := 3

	original := CompletionRequest{
		Model:            "gpt-3.5-turbo-instruct",
		Prompt:           "Say hello",
		MaxTokens:        &maxTk,
		Temperature:      &temp,
		TopP:             &temp,
		Stream:           true,
		Stop:             json.RawMessage(`["END"]`),
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.2,
		User:             "user-123",
		Echo:             &echo,
		Logprobs:         &logprobs,
		BestOf:           &bestOf,
		Suffix:           "!",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CompletionRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Model != original.Model {
		t.Errorf("model = %q, want %q", decoded.Model, original.Model)
	}
	if decoded.Prompt != original.Prompt {
		t.Errorf("prompt = %q, want %q", decoded.Prompt, original.Prompt)
	}
	if decoded.MaxTokens == nil || *decoded.MaxTokens != maxTk {
		t.Errorf("max_tokens = %v, want %v", decoded.MaxTokens, maxTk)
	}
	if !decoded.Stream {
		t.Error("stream should be true")
	}
	if decoded.Echo == nil || *decoded.Echo != echo {
		t.Errorf("echo = %v, want %v", decoded.Echo, echo)
	}
	if decoded.Logprobs == nil || *decoded.Logprobs != logprobs {
		t.Errorf("logprobs = %v, want %v", decoded.Logprobs, logprobs)
	}
	if decoded.BestOf == nil || *decoded.BestOf != bestOf {
		t.Errorf("best_of = %v, want %v", decoded.BestOf, bestOf)
	}
}

func TestMessage_ContentString_StringValue(t *testing.T) {
	m := Message{Role: "user", Content: StringContent("hello")}
	if m.ContentString() != "hello" {
		t.Errorf("ContentString() = %q, want hello", m.ContentString())
	}
}

func TestMessage_ContentParts_StringValue(t *testing.T) {
	m := Message{Role: "user", Content: StringContent("hello")}
	parts, err := m.ContentParts()
	if err != nil {
		t.Fatalf("ContentParts() error = %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("len(ContentParts()) = %d, want 1", len(parts))
	}
	if parts[0].Type != "text" || parts[0].Text != "hello" {
		t.Errorf("ContentParts()[0] = %+v, want {Type:text Text:hello}", parts[0])
	}
}

func TestMessage_ContentParts_ArrayValue(t *testing.T) {
	m := Message{
		Role: "user",
		Content: json.RawMessage(
			`[{"type":"text","text":"look"},{"type":"image_url","image_url":{"url":"https://example.com/cat.png","detail":"high"}}]`,
		),
	}

	parts, err := m.ContentParts()
	if err != nil {
		t.Fatalf("ContentParts() error = %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("len(ContentParts()) = %d, want 2", len(parts))
	}
	if parts[0].Type != "text" || parts[0].Text != "look" {
		t.Errorf("parts[0] = %+v, want text part", parts[0])
	}
	if parts[1].Type != "image_url" || parts[1].ImageURL == nil || parts[1].ImageURL.URL != "https://example.com/cat.png" {
		t.Errorf("parts[1] = %+v, want image_url part", parts[1])
	}
	if parts[1].ImageURL.Detail != "high" {
		t.Errorf("parts[1].image_url.detail = %q, want high", parts[1].ImageURL.Detail)
	}
}

func TestMessage_ContentParts_InvalidValue(t *testing.T) {
	m := Message{Role: "user", Content: json.RawMessage(`123`)}
	if _, err := m.ContentParts(); err == nil {
		t.Fatal("ContentParts() expected error for invalid content type")
	}
}

func TestMessage_ContentString_ArrayValue(t *testing.T) {
	// Array content should concatenate text parts.
	m := Message{Role: "user", Content: json.RawMessage(`[{"type":"text","text":"hi"},{"type":"image_url","image_url":{"url":"..."}},{"type":"text","text":" there"}]`)}
	if m.ContentString() != "hi there" {
		t.Errorf("ContentString() for array = %q, want %q", m.ContentString(), "hi there")
	}
}

func TestMessage_ContentString_ImageOnlyArrayValue(t *testing.T) {
	m := Message{Role: "user", Content: json.RawMessage(`[{"type":"image_url","image_url":{"url":"https://example.com/cat.png"}}]`)}
	if m.ContentString() != "" {
		t.Errorf("ContentString() for image-only content = %q, want empty", m.ContentString())
	}
}

func TestMessage_ContentString_NilContent(t *testing.T) {
	m := Message{Role: "user"}
	if m.ContentString() != "" {
		t.Errorf("ContentString() for nil content = %q, want empty", m.ContentString())
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
				Message:      &Message{Role: "assistant", Content: StringContent("Hello!")},
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
	if decoded.Choices[0].Message.ContentString() != "Hello!" {
		t.Errorf("content = %q, want Hello!", decoded.Choices[0].Message.ContentString())
	}
	if decoded.Usage.TotalTokens != 15 {
		t.Errorf("total_tokens = %d, want 15", decoded.Usage.TotalTokens)
	}
}

func TestCompletionResponse_JSONRoundTrip(t *testing.T) {
	original := CompletionResponse{
		ID:      "cmpl-abc123",
		Object:  "text_completion",
		Created: 1700000000,
		Model:   "gpt-3.5-turbo-instruct",
		Choices: []CompletionChoice{
			{
				Text:         "Hello!",
				Index:        0,
				FinishReason: "stop",
			},
		},
		Usage: &Usage{
			PromptTokens:     3,
			CompletionTokens: 1,
			TotalTokens:      4,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CompletionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Object != "text_completion" {
		t.Errorf("object = %q, want text_completion", decoded.Object)
	}
	if len(decoded.Choices) != 1 {
		t.Fatalf("choices length = %d, want 1", len(decoded.Choices))
	}
	if decoded.Choices[0].Text != "Hello!" {
		t.Errorf("choices[0].text = %q, want Hello!", decoded.Choices[0].Text)
	}
	if decoded.Usage == nil || decoded.Usage.TotalTokens != 4 {
		t.Errorf("usage.total_tokens = %v, want 4", decoded.Usage)
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
				Delta: &MessageDelta{Content: StringContent("Hello")},
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
	if decoded.Choices[0].Delta.ContentString() != "Hello" {
		t.Errorf("delta.content = %q, want Hello", decoded.Choices[0].Delta.ContentString())
	}
}

func TestChatCompletionRequest_ToolFields(t *testing.T) {
	raw := `{
		"model": "gpt-4o",
		"messages": [{"role":"user","content":"hi"}],
		"tools": [{"type":"function","function":{"name":"get_weather","parameters":{"type":"object"}}}],
		"tool_choice": "auto"
	}`

	var req ChatCompletionRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(req.Tools) != 1 {
		t.Fatalf("tools count = %d, want 1", len(req.Tools))
	}
	if req.Tools[0].Type != "function" {
		t.Errorf("tool type = %q, want function", req.Tools[0].Type)
	}
	if req.Tools[0].Function.Name != "get_weather" {
		t.Errorf("function name = %q, want get_weather", req.Tools[0].Function.Name)
	}
	if req.ToolChoice == nil {
		t.Error("tool_choice should not be nil")
	}
	var tc string
	if err := json.Unmarshal(req.ToolChoice, &tc); err != nil || tc != "auto" {
		t.Errorf("tool_choice = %q, want auto", tc)
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
