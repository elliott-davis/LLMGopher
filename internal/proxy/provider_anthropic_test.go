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
					{Role: "system", Content: "You are helpful."},
					{Role: "user", Content: "Hello"},
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
					{Role: "user", Content: "Hi"},
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
					{Role: "system", Content: "Be concise."},
					{Role: "user", Content: "What is Go?"},
					{Role: "assistant", Content: "A programming language."},
					{Role: "user", Content: "Tell me more."},
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
				Messages: []llm.Message{{Role: "user", Content: "hi"}},
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
			{Role: "user", Content: "question"},
			{Role: "assistant", Content: "answer"},
			{Role: "user", Content: "follow-up"},
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
			if choice.Message.Content != tc.wantText {
				t.Errorf("content = %q, want %q", choice.Message.Content, tc.wantText)
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

func TestMapAnthropicStopReason(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"end_turn", "stop"},
		{"max_tokens", "length"},
		{"stop_sequence", "stop"},
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
