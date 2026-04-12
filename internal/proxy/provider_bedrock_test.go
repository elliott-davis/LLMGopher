package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockdocument "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	bedrocktypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/ed007183/llmgopher/pkg/llm"
)

type mockBedrockClient struct {
	converseInput *bedrockruntime.ConverseInput
	converseFn    func(*bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error)

	streamInput *bedrockruntime.ConverseStreamInput
	streamFn    func(*bedrockruntime.ConverseStreamInput) (bedrockConverseStream, error)
}

func (m *mockBedrockClient) Converse(_ context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
	m.converseInput = input
	if m.converseFn != nil {
		return m.converseFn(input)
	}
	return &bedrockruntime.ConverseOutput{}, nil
}

func (m *mockBedrockClient) ConverseStream(_ context.Context, input *bedrockruntime.ConverseStreamInput) (bedrockConverseStream, error) {
	m.streamInput = input
	if m.streamFn != nil {
		return m.streamFn(input)
	}
	return &mockBedrockStream{}, nil
}

type mockBedrockStream struct {
	events chan bedrocktypes.ConverseStreamOutput
	err    error
	closed bool
}

func (s *mockBedrockStream) Events() <-chan bedrocktypes.ConverseStreamOutput {
	if s.events == nil {
		ch := make(chan bedrocktypes.ConverseStreamOutput)
		close(ch)
		return ch
	}
	return s.events
}

func (s *mockBedrockStream) Err() error {
	return s.err
}

func (s *mockBedrockStream) Close() error {
	s.closed = true
	return nil
}

func TestBedrockProviderChatCompletion_TranslatesRequestAndResponse(t *testing.T) {
	t.Parallel()

	mockClient := &mockBedrockClient{
		converseFn: func(input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return &bedrockruntime.ConverseOutput{
				Output: &bedrocktypes.ConverseOutputMemberMessage{
					Value: bedrocktypes.Message{
						Role: bedrocktypes.ConversationRoleAssistant,
						Content: []bedrocktypes.ContentBlock{
							&bedrocktypes.ContentBlockMemberText{Value: "Hello from Bedrock"},
						},
					},
				},
				StopReason: bedrocktypes.StopReasonEndTurn,
				Usage: &bedrocktypes.TokenUsage{
					InputTokens:  i32ptr(12),
					OutputTokens: i32ptr(7),
					TotalTokens:  i32ptr(19),
				},
			}, nil
		},
	}

	provider := &BedrockProvider{
		client: mockClient,
		region: "us-east-1",
	}

	stop := json.RawMessage(`"END"`)
	response, err := provider.ChatCompletion(t.Context(), &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet",
		Messages: []llm.Message{
			{Role: "system", Content: llm.StringContent("be concise")},
			{Role: "user", Content: llm.StringContent("hello")},
		},
		Stop:  stop,
		Tools: []llm.Tool{{Type: "function", Function: llm.Function{Name: "weather", Parameters: json.RawMessage(`{"type":"object"}`)}}},
	})
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v, want nil", err)
	}

	if mockClient.converseInput == nil {
		t.Fatal("expected Converse input to be captured")
	}
	if gotModel := derefString(mockClient.converseInput.ModelId); gotModel != BedrockModelAliases["claude-3-5-sonnet"] {
		t.Fatalf("ModelId = %q, want alias %q", gotModel, BedrockModelAliases["claude-3-5-sonnet"])
	}
	if mockClient.converseInput.InferenceConfig == nil || derefInt32(mockClient.converseInput.InferenceConfig.MaxTokens) != bedrockDefaultMaxTokens {
		t.Fatalf("max tokens = %d, want %d", derefInt32(mockClient.converseInput.InferenceConfig.MaxTokens), bedrockDefaultMaxTokens)
	}
	if len(mockClient.converseInput.System) != 1 {
		t.Fatalf("system blocks = %d, want 1", len(mockClient.converseInput.System))
	}
	if len(mockClient.converseInput.Messages) != 1 {
		t.Fatalf("messages = %d, want 1 (system extracted)", len(mockClient.converseInput.Messages))
	}
	if len(mockClient.converseInput.InferenceConfig.StopSequences) != 1 || mockClient.converseInput.InferenceConfig.StopSequences[0] != "END" {
		t.Fatalf("stop sequences = %v, want [END]", mockClient.converseInput.InferenceConfig.StopSequences)
	}
	if mockClient.converseInput.ToolConfig == nil || len(mockClient.converseInput.ToolConfig.Tools) != 1 {
		t.Fatalf("tool config not translated: %+v", mockClient.converseInput.ToolConfig)
	}

	if response.Choices[0].Message == nil || response.Choices[0].Message.ContentString() != "Hello from Bedrock" {
		t.Fatalf("response content = %q, want %q", response.Choices[0].Message.ContentString(), "Hello from Bedrock")
	}
	if response.Choices[0].FinishReason != "stop" {
		t.Fatalf("finish reason = %q, want stop", response.Choices[0].FinishReason)
	}
	if response.Usage == nil || response.Usage.PromptTokens != 12 || response.Usage.CompletionTokens != 7 || response.Usage.TotalTokens != 19 {
		t.Fatalf("usage = %+v, want {12,7,19}", response.Usage)
	}
}

func TestBedrockProviderChatCompletion_TranslatesToolUseResponse(t *testing.T) {
	t.Parallel()

	mockClient := &mockBedrockClient{
		converseFn: func(_ *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return &bedrockruntime.ConverseOutput{
				Output: &bedrocktypes.ConverseOutputMemberMessage{
					Value: bedrocktypes.Message{
						Role: bedrocktypes.ConversationRoleAssistant,
						Content: []bedrocktypes.ContentBlock{
							&bedrocktypes.ContentBlockMemberToolUse{
								Value: bedrocktypes.ToolUseBlock{
									ToolUseId: strptr("tool_1"),
									Name:      strptr("lookup_weather"),
									Input: bedrockdocument.NewLazyDocument(map[string]any{
										"city": "Seattle",
									}),
								},
							},
						},
					},
				},
				StopReason: bedrocktypes.StopReasonToolUse,
				Usage: &bedrocktypes.TokenUsage{
					InputTokens:  i32ptr(5),
					OutputTokens: i32ptr(4),
					TotalTokens:  i32ptr(9),
				},
			}, nil
		},
	}

	provider := &BedrockProvider{client: mockClient, region: "us-east-1"}
	resp, err := provider.ChatCompletion(t.Context(), &llm.ChatCompletionRequest{
		Model:    "anthropic.claude-3-5-sonnet-20241022-v2:0",
		Messages: []llm.Message{{Role: "user", Content: llm.StringContent("weather?")}},
	})
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v, want nil", err)
	}

	if resp.Choices[0].FinishReason != "tool_calls" {
		t.Fatalf("finish reason = %q, want tool_calls", resp.Choices[0].FinishReason)
	}
	if resp.Choices[0].Message == nil || len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("tool calls = %+v, want 1", resp.Choices[0].Message)
	}
	toolCall := resp.Choices[0].Message.ToolCalls[0]
	if toolCall.ID != "tool_1" || toolCall.Function.Name != "lookup_weather" {
		t.Fatalf("tool call = %+v, want id tool_1 name lookup_weather", toolCall)
	}
	if !strings.Contains(toolCall.Function.Arguments, `"Seattle"`) {
		t.Fatalf("tool call arguments = %q, want Seattle", toolCall.Function.Arguments)
	}
}

func TestBedrockProviderChatCompletionStream_TranslatesEventsToSSE(t *testing.T) {
	t.Parallel()

	eventCh := make(chan bedrocktypes.ConverseStreamOutput, 6)
	eventCh <- &bedrocktypes.ConverseStreamOutputMemberContentBlockStart{
		Value: bedrocktypes.ContentBlockStartEvent{
			ContentBlockIndex: i32ptr(0),
			Start: &bedrocktypes.ContentBlockStartMemberToolUse{
				Value: bedrocktypes.ToolUseBlockStart{
					ToolUseId: strptr("tool_stream_1"),
					Name:      strptr("lookup_weather"),
				},
			},
		},
	}
	eventCh <- &bedrocktypes.ConverseStreamOutputMemberContentBlockDelta{
		Value: bedrocktypes.ContentBlockDeltaEvent{
			ContentBlockIndex: i32ptr(0),
			Delta: &bedrocktypes.ContentBlockDeltaMemberToolUse{
				Value: bedrocktypes.ToolUseBlockDelta{
					Input: strptr(`{"city":"`),
				},
			},
		},
	}
	eventCh <- &bedrocktypes.ConverseStreamOutputMemberContentBlockDelta{
		Value: bedrocktypes.ContentBlockDeltaEvent{
			ContentBlockIndex: i32ptr(0),
			Delta: &bedrocktypes.ContentBlockDeltaMemberToolUse{
				Value: bedrocktypes.ToolUseBlockDelta{
					Input: strptr(`Seattle"}`),
				},
			},
		},
	}
	eventCh <- &bedrocktypes.ConverseStreamOutputMemberMetadata{
		Value: bedrocktypes.ConverseStreamMetadataEvent{
			Usage: &bedrocktypes.TokenUsage{
				InputTokens:  i32ptr(14),
				OutputTokens: i32ptr(8),
				TotalTokens:  i32ptr(22),
			},
		},
	}
	eventCh <- &bedrocktypes.ConverseStreamOutputMemberMessageStop{
		Value: bedrocktypes.MessageStopEvent{
			StopReason: bedrocktypes.StopReasonToolUse,
		},
	}
	close(eventCh)

	stream := &mockBedrockStream{events: eventCh}
	mockClient := &mockBedrockClient{
		streamFn: func(_ *bedrockruntime.ConverseStreamInput) (bedrockConverseStream, error) {
			return stream, nil
		},
	}
	provider := &BedrockProvider{client: mockClient, region: "us-east-1"}

	reader, err := provider.ChatCompletionStream(t.Context(), &llm.ChatCompletionRequest{
		Model:    "claude-3-5-sonnet",
		Stream:   true,
		Messages: []llm.Message{{Role: "user", Content: llm.StringContent("weather")}},
	})
	if err != nil {
		t.Fatalf("ChatCompletionStream() error = %v, want nil", err)
	}
	defer reader.Close()

	dataEvents := make([]string, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			dataEvents = append(dataEvents, strings.TrimPrefix(line, "data: "))
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		t.Fatalf("stream read error: %v", err)
	}
	if len(dataEvents) == 0 {
		t.Fatal("expected translated SSE events")
	}
	if dataEvents[len(dataEvents)-1] != "[DONE]" {
		t.Fatalf("last SSE event = %q, want [DONE]", dataEvents[len(dataEvents)-1])
	}

	var chunks []llm.ChatCompletionChunk
	for _, event := range dataEvents[:len(dataEvents)-1] {
		var chunk llm.ChatCompletionChunk
		if err := json.Unmarshal([]byte(event), &chunk); err != nil {
			t.Fatalf("unmarshal chunk %q: %v", event, err)
		}
		chunks = append(chunks, chunk)
	}

	foundStart := false
	foundFinish := false
	argBuilder := strings.Builder{}
	var usage *llm.Usage
	for _, chunk := range chunks {
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		for _, choice := range chunk.Choices {
			if choice.FinishReason == "tool_calls" {
				foundFinish = true
			}
			if choice.Delta == nil {
				continue
			}
			for _, toolCall := range choice.Delta.ToolCalls {
				if toolCall.ID == "tool_stream_1" && toolCall.Function.Name == "lookup_weather" {
					foundStart = true
				}
				argBuilder.WriteString(toolCall.Function.Arguments)
			}
		}
	}

	if !foundStart {
		t.Fatal("expected tool-call start chunk with id/name")
	}
	if argBuilder.String() != `{"city":"Seattle"}` {
		t.Fatalf("tool-call arguments = %q, want %q", argBuilder.String(), `{"city":"Seattle"}`)
	}
	if !foundFinish {
		t.Fatal("expected finish_reason=tool_calls chunk")
	}
	if usage == nil || usage.PromptTokens != 14 || usage.CompletionTokens != 8 || usage.TotalTokens != 22 {
		t.Fatalf("usage = %+v, want {14,8,22}", usage)
	}
	if !stream.closed {
		t.Fatal("expected source stream Close() to be called")
	}
}

func TestResolveBedrockRegion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "region only", input: "us-west-2", want: "us-west-2"},
		{name: "runtime url", input: "https://bedrock-runtime.us-east-1.amazonaws.com", want: "us-east-1"},
		{name: "runtime host", input: "bedrock-runtime.eu-central-1.amazonaws.com", want: "eu-central-1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveBedrockRegion(tc.input)
			if got != tc.want {
				t.Fatalf("resolveBedrockRegion(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func i32ptr(v int32) *int32   { return &v }
func strptr(v string) *string { return &v }

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}
