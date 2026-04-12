package proxy

import (
	"encoding/json"
	"testing"

	"github.com/ed007183/llmgopher/pkg/llm"
)

func ptr[T any](v T) *T { return &v }

func TestToAnthropicRequest_BasicMapping(t *testing.T) {
	p := &AnthropicProvider{}

	tests := []struct {
		name     string
		input    *llm.ChatCompletionRequest
		wantMsgs int
		wantSys  string
		wantMax  int
	}{
		{
			name: "system message extracted",
			input: &llm.ChatCompletionRequest{
				Model: "claude-3-5-sonnet-20241022",
				Messages: []llm.Message{
					{Role: "system", Content: llm.StringContent("You are helpful.")},
					{Role: "user", Content: llm.StringContent("Hello")},
				},
				MaxTokens: ptr(512),
			},
			wantMsgs: 1,
			wantSys:  "You are helpful.",
			wantMax:  512,
		},
		{
			name: "no system message",
			input: &llm.ChatCompletionRequest{
				Model: "claude-3-haiku-20240307",
				Messages: []llm.Message{
					{Role: "user", Content: llm.StringContent("Hi")},
				},
			},
			wantMsgs: 1,
			wantSys:  "",
			wantMax:  4096,
		},
		{
			name: "multi-turn conversation",
			input: &llm.ChatCompletionRequest{
				Model: "claude-3-opus-20240229",
				Messages: []llm.Message{
					{Role: "system", Content: llm.StringContent("Be concise.")},
					{Role: "user", Content: llm.StringContent("What is Go?")},
					{Role: "assistant", Content: llm.StringContent("A programming language.")},
					{Role: "user", Content: llm.StringContent("Tell me more.")},
				},
				Temperature: ptr(0.7),
			},
			wantMsgs: 3,
			wantSys:  "Be concise.",
			wantMax:  4096,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ar := p.toAnthropicRequest(tc.input)

			if len(ar.Messages) != tc.wantMsgs {
				t.Errorf("messages count = %d, want %d", len(ar.Messages), tc.wantMsgs)
			}
			if ar.System != tc.wantSys {
				t.Errorf("system = %q, want %q", ar.System, tc.wantSys)
			}
			if ar.MaxTokens != tc.wantMax {
				t.Errorf("max_tokens = %d, want %d", ar.MaxTokens, tc.wantMax)
			}
			if ar.Model != tc.input.Model {
				t.Errorf("model = %q, want %q", ar.Model, tc.input.Model)
			}
		})
	}
}

func TestToAnthropicRequest_StopSequences(t *testing.T) {
	p := &AnthropicProvider{}

	tests := []struct {
		name     string
		stop     json.RawMessage
		wantStop []string
	}{
		{
			name:     "string array stop",
			stop:     json.RawMessage(`["END","STOP"]`),
			wantStop: []string{"END", "STOP"},
		},
		{
			name:     "single string stop",
			stop:     json.RawMessage(`"END"`),
			wantStop: []string{"END"},
		},
		{
			name:     "nil stop",
			stop:     nil,
			wantStop: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &llm.ChatCompletionRequest{
				Model:    "claude-3-haiku-20240307",
				Messages: []llm.Message{{Role: "user", Content: llm.StringContent("hi")}},
				Stop:     tc.stop,
			}
			ar := p.toAnthropicRequest(req)

			if len(ar.StopSeqs) != len(tc.wantStop) {
				t.Fatalf("stop_sequences length = %d, want %d", len(ar.StopSeqs), len(tc.wantStop))
			}
			for i := range tc.wantStop {
				if ar.StopSeqs[i] != tc.wantStop[i] {
					t.Errorf("stop_sequences[%d] = %q, want %q", i, ar.StopSeqs[i], tc.wantStop[i])
				}
			}
		})
	}
}

func TestToAnthropicRequest_RoleMapping(t *testing.T) {
	p := &AnthropicProvider{}

	req := &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: llm.StringContent("question")},
			{Role: "assistant", Content: llm.StringContent("answer")},
			{Role: "user", Content: llm.StringContent("follow-up")},
		},
	}

	ar := p.toAnthropicRequest(req)

	wantRoles := []string{"user", "assistant", "user"}
	for i, want := range wantRoles {
		if ar.Messages[i].Role != want {
			t.Errorf("messages[%d].role = %q, want %q", i, ar.Messages[i].Role, want)
		}
	}
}

func TestToAnthropicRequest_VisionURLContentParts(t *testing.T) {
	p := &AnthropicProvider{}
	req := &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{
				Role: "user",
				Content: json.RawMessage(
					`[{"type":"text","text":"what is in this image?"},{"type":"image_url","image_url":{"url":"https://example.com/cat.png"}}]`,
				),
			},
		},
	}

	ar := p.toAnthropicRequest(req)
	if len(ar.Messages) != 1 {
		t.Fatalf("messages count = %d, want 1", len(ar.Messages))
	}

	var blocks []anthropicContentBlock
	if err := json.Unmarshal(ar.Messages[0].Content, &blocks); err != nil {
		t.Fatalf("unmarshal anthropic content blocks: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("content blocks count = %d, want 2", len(blocks))
	}
	if blocks[0].Type != "text" || blocks[0].Text != "what is in this image?" {
		t.Errorf("blocks[0] = %+v, want text block", blocks[0])
	}
	if blocks[1].Type != "image" || blocks[1].Source == nil {
		t.Fatalf("blocks[1] = %+v, want image block with source", blocks[1])
	}
	if blocks[1].Source.Type != "url" || blocks[1].Source.URL != "https://example.com/cat.png" {
		t.Errorf("blocks[1].source = %+v, want url source", blocks[1].Source)
	}
}

func TestToAnthropicRequest_VisionDataURIContentPart(t *testing.T) {
	p := &AnthropicProvider{}
	req := &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{
				Role: "user",
				Content: json.RawMessage(
					`[{"type":"image_url","image_url":{"url":"data:image/png;base64,aGVsbG8="}}]`,
				),
			},
		},
	}

	ar := p.toAnthropicRequest(req)
	if len(ar.Messages) != 1 {
		t.Fatalf("messages count = %d, want 1", len(ar.Messages))
	}

	var blocks []anthropicContentBlock
	if err := json.Unmarshal(ar.Messages[0].Content, &blocks); err != nil {
		t.Fatalf("unmarshal anthropic content blocks: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("content blocks count = %d, want 1", len(blocks))
	}
	if blocks[0].Type != "image" || blocks[0].Source == nil {
		t.Fatalf("blocks[0] = %+v, want image block with source", blocks[0])
	}
	if blocks[0].Source.Type != "base64" {
		t.Errorf("source.type = %q, want base64", blocks[0].Source.Type)
	}
	if blocks[0].Source.MediaType != "image/png" {
		t.Errorf("source.media_type = %q, want image/png", blocks[0].Source.MediaType)
	}
	if blocks[0].Source.Data != "aGVsbG8=" {
		t.Errorf("source.data = %q, want aGVsbG8=", blocks[0].Source.Data)
	}
}

func TestToAnthropicRequest_ToolTranslation(t *testing.T) {
	p := &AnthropicProvider{}

	req := &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: llm.StringContent("What is the weather?")},
		},
		Tools: []llm.Tool{
			{
				Type: "function",
				Function: llm.Function{
					Name:        "get_weather",
					Description: "Get current weather",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"location":{"type":"string"}},"required":["location"]}`),
				},
			},
		},
		ToolChoice: json.RawMessage(`"auto"`),
	}

	ar := p.toAnthropicRequest(req)

	if len(ar.Tools) != 1 {
		t.Fatalf("tools count = %d, want 1", len(ar.Tools))
	}
	if ar.Tools[0].Name != "get_weather" {
		t.Errorf("tool name = %q, want get_weather", ar.Tools[0].Name)
	}
	if ar.Tools[0].Description != "Get current weather" {
		t.Errorf("tool description = %q, want Get current weather", ar.Tools[0].Description)
	}
	if ar.Tools[0].InputSchema == nil {
		t.Error("tool input_schema is nil")
	}
	if ar.ToolChoice == nil {
		t.Fatal("tool_choice is nil")
	}
	if ar.ToolChoice.Type != "auto" {
		t.Errorf("tool_choice.type = %q, want auto", ar.ToolChoice.Type)
	}
}

func TestParseAnthropicToolChoice(t *testing.T) {
	tests := []struct {
		name     string
		input    json.RawMessage
		wantType string
		wantName string
	}{
		{"none", json.RawMessage(`"none"`), "none", ""},
		{"auto", json.RawMessage(`"auto"`), "auto", ""},
		{"required→any", json.RawMessage(`"required"`), "any", ""},
		{"function object", json.RawMessage(`{"type":"function","function":{"name":"get_weather"}}`), "tool", "get_weather"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseAnthropicToolChoice(tc.input)
			if got == nil {
				t.Fatal("result is nil")
			}
			if got.Type != tc.wantType {
				t.Errorf("type = %q, want %q", got.Type, tc.wantType)
			}
			if got.Name != tc.wantName {
				t.Errorf("name = %q, want %q", got.Name, tc.wantName)
			}
		})
	}
}

func TestToAnthropicRequest_ToolRoleMessage(t *testing.T) {
	p := &AnthropicProvider{}

	req := &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: llm.StringContent("What is the weather?")},
			{
				Role: "assistant",
				ToolCalls: []llm.ToolCall{
					{ID: "call_123", Type: "function", Function: llm.FunctionCall{Name: "get_weather", Arguments: `{"location":"NYC"}`}},
				},
			},
			{Role: "tool", Content: llm.StringContent(`{"temp":72}`), ToolCallID: "call_123"},
		},
	}

	ar := p.toAnthropicRequest(req)

	// user + assistant (with tool_use) + user (with tool_result) = 3 messages
	if len(ar.Messages) != 3 {
		t.Fatalf("messages count = %d, want 3", len(ar.Messages))
	}

	// The last message should be a user message with tool_result content.
	lastMsg := ar.Messages[2]
	if lastMsg.Role != "user" {
		t.Errorf("last message role = %q, want user", lastMsg.Role)
	}
	var blocks []anthropicContentBlock
	if err := json.Unmarshal(lastMsg.Content, &blocks); err != nil {
		t.Fatalf("unmarshal last message content: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Type != "tool_result" {
		t.Errorf("last message blocks = %+v, want single tool_result", blocks)
	}
	if blocks[0].ToolUseID != "call_123" {
		t.Errorf("tool_use_id = %q, want call_123", blocks[0].ToolUseID)
	}
}

func TestFromAnthropicResponse(t *testing.T) {
	p := &AnthropicProvider{}

	tests := []struct {
		name         string
		resp         *anthropicResponse
		model        string
		wantText     string
		wantFinish   string
		wantPromptTk int
		wantOutTk    int
	}{
		{
			name: "normal completion",
			resp: &anthropicResponse{
				ID:   "msg_123",
				Type: "message",
				Role: "assistant",
				Content: []anthropicContentBlock{
					{Type: "text", Text: "Hello, world!"},
				},
				StopReason: "end_turn",
				Usage:      anthropicUsage{InputTokens: 10, OutputTokens: 5},
			},
			model:        "claude-3-5-sonnet-20241022",
			wantText:     "Hello, world!",
			wantFinish:   "stop",
			wantPromptTk: 10,
			wantOutTk:    5,
		},
		{
			name: "max_tokens stop",
			resp: &anthropicResponse{
				ID:   "msg_456",
				Type: "message",
				Role: "assistant",
				Content: []anthropicContentBlock{
					{Type: "text", Text: "Truncated output"},
				},
				StopReason: "max_tokens",
				Usage:      anthropicUsage{InputTokens: 20, OutputTokens: 100},
			},
			model:        "claude-3-opus-20240229",
			wantText:     "Truncated output",
			wantFinish:   "length",
			wantPromptTk: 20,
			wantOutTk:    100,
		},
		{
			name: "multiple content blocks concatenated",
			resp: &anthropicResponse{
				ID:   "msg_789",
				Type: "message",
				Role: "assistant",
				Content: []anthropicContentBlock{
					{Type: "text", Text: "Part 1. "},
					{Type: "text", Text: "Part 2."},
				},
				StopReason: "end_turn",
				Usage:      anthropicUsage{InputTokens: 5, OutputTokens: 8},
			},
			model:    "claude-3-haiku-20240307",
			wantText: "Part 1. Part 2.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := p.fromAnthropicResponse(tc.resp, tc.model)

			if resp.Object != "chat.completion" {
				t.Errorf("object = %q, want %q", resp.Object, "chat.completion")
			}
			if resp.Model != tc.model {
				t.Errorf("model = %q, want %q", resp.Model, tc.model)
			}
			if len(resp.Choices) != 1 {
				t.Fatalf("choices length = %d, want 1", len(resp.Choices))
			}

			choice := resp.Choices[0]
			if choice.Message == nil {
				t.Fatal("choice.message is nil")
			}
			if choice.Message.ContentString() != tc.wantText {
				t.Errorf("content = %q, want %q", choice.Message.ContentString(), tc.wantText)
			}
			if tc.wantFinish != "" && choice.FinishReason != tc.wantFinish {
				t.Errorf("finish_reason = %q, want %q", choice.FinishReason, tc.wantFinish)
			}
			if tc.wantPromptTk > 0 && resp.Usage.PromptTokens != tc.wantPromptTk {
				t.Errorf("prompt_tokens = %d, want %d", resp.Usage.PromptTokens, tc.wantPromptTk)
			}
			if tc.wantOutTk > 0 && resp.Usage.CompletionTokens != tc.wantOutTk {
				t.Errorf("completion_tokens = %d, want %d", resp.Usage.CompletionTokens, tc.wantOutTk)
			}
		})
	}
}

func TestFromAnthropicResponse_ToolUse(t *testing.T) {
	p := &AnthropicProvider{}

	resp := &anthropicResponse{
		ID:   "msg_tool_123",
		Type: "message",
		Role: "assistant",
		Content: []anthropicContentBlock{
			{
				Type:  "tool_use",
				ID:    "toolu_01abc",
				Name:  "get_weather",
				Input: json.RawMessage(`{"location":"San Francisco","unit":"celsius"}`),
			},
		},
		StopReason: "tool_use",
		Usage:      anthropicUsage{InputTokens: 15, OutputTokens: 20},
	}

	out := p.fromAnthropicResponse(resp, "claude-3-5-sonnet-20241022")

	if out.Choices[0].FinishReason != "tool_calls" {
		t.Errorf("finish_reason = %q, want tool_calls", out.Choices[0].FinishReason)
	}
	msg := out.Choices[0].Message
	if msg == nil {
		t.Fatal("message is nil")
	}
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("tool_calls count = %d, want 1", len(msg.ToolCalls))
	}
	tc := msg.ToolCalls[0]
	if tc.ID != "toolu_01abc" {
		t.Errorf("tool_call.id = %q, want toolu_01abc", tc.ID)
	}
	if tc.Type != "function" {
		t.Errorf("tool_call.type = %q, want function", tc.Type)
	}
	if tc.Function.Name != "get_weather" {
		t.Errorf("tool_call.function.name = %q, want get_weather", tc.Function.Name)
	}
	// Arguments should be the JSON-encoded input.
	var args map[string]string
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		t.Fatalf("unmarshal arguments: %v", err)
	}
	if args["location"] != "San Francisco" {
		t.Errorf("arguments.location = %q, want San Francisco", args["location"])
	}
}

func TestMapAnthropicStopReason(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"end_turn", "stop"},
		{"max_tokens", "length"},
		{"stop_sequence", "stop"},
		{"tool_use", "tool_calls"},
		{"unknown_reason", "unknown_reason"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := mapAnthropicStopReason(tc.input)
			if got != tc.want {
				t.Errorf("mapAnthropicStopReason(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
