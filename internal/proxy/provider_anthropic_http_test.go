package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// fakeAnthropicServer starts an httptest.Server that responds with a
// pre-canned Anthropic Messages API response body and status code.
func fakeAnthropicServer(t *testing.T, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprint(w, body)
	}))
}

// fakeAnthropicSSEServer starts an httptest.Server that streams pre-canned
// Anthropic SSE events.
func fakeAnthropicSSEServer(t *testing.T, events []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		for _, e := range events {
			fmt.Fprintf(w, "data: %s\n\n", e)
		}
	}))
}

// --------------------------------------------------------------------------
// Sync tool-call round-trip
// --------------------------------------------------------------------------

func TestAnthropicProvider_ToolCall_Sync(t *testing.T) {
	// Simulate Anthropic returning a tool_use content block.
	anthropicResp := `{
		"id": "msg_tool_001",
		"type": "message",
		"role": "assistant",
		"content": [
			{
				"type": "tool_use",
				"id": "toolu_01XYZ",
				"name": "get_weather",
				"input": {"location": "Paris", "unit": "celsius"}
			}
		],
		"model": "claude-3-5-sonnet-20241022",
		"stop_reason": "tool_use",
		"usage": {"input_tokens": 25, "output_tokens": 15}
	}`

	srv := fakeAnthropicServer(t, http.StatusOK, anthropicResp)
	defer srv.Close()

	p := NewAnthropicProvider("test-key", srv.URL)

	req := &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: llm.StringContent("What is the weather in Paris?")},
		},
		Tools: []llm.Tool{{
			Type: "function",
			Function: llm.Function{
				Name:        "get_weather",
				Description: "Get current weather",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"location":{"type":"string"},"unit":{"type":"string"}},"required":["location"]}`),
			},
		}},
		ToolChoice: json.RawMessage(`"auto"`),
	}

	resp, err := p.ChatCompletion(t.Context(), req)
	if err != nil {
		t.Fatalf("ChatCompletion error: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("choices count = %d, want 1", len(resp.Choices))
	}
	choice := resp.Choices[0]

	if choice.FinishReason != "tool_calls" {
		t.Errorf("finish_reason = %q, want tool_calls", choice.FinishReason)
	}

	msg := choice.Message
	if msg == nil {
		t.Fatal("message is nil")
	}
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("tool_calls count = %d, want 1", len(msg.ToolCalls))
	}

	tc := msg.ToolCalls[0]
	if tc.ID != "toolu_01XYZ" {
		t.Errorf("id = %q, want toolu_01XYZ", tc.ID)
	}
	if tc.Type != "function" {
		t.Errorf("type = %q, want function", tc.Type)
	}
	if tc.Function.Name != "get_weather" {
		t.Errorf("function.name = %q, want get_weather", tc.Function.Name)
	}

	var args map[string]string
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		t.Fatalf("unmarshal arguments: %v", err)
	}
	if args["location"] != "Paris" {
		t.Errorf("arguments.location = %q, want Paris", args["location"])
	}
	if args["unit"] != "celsius" {
		t.Errorf("arguments.unit = %q, want celsius", args["unit"])
	}

	// Verify usage was translated.
	if resp.Usage == nil || resp.Usage.PromptTokens != 25 || resp.Usage.CompletionTokens != 15 {
		t.Errorf("usage = %+v, want {25, 15, 40}", resp.Usage)
	}
}

// TestAnthropicProvider_ToolCall_RequestWireFormat verifies that the outbound
// request to Anthropic has the correct wire format: tools array with
// input_schema, and tool_choice translated to Anthropic's format.
func TestAnthropicProvider_ToolCall_RequestWireFormat(t *testing.T) {
	var captured anthropicRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		// Return a minimal valid response.
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"msg_x","type":"message","role":"assistant","content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":5,"output_tokens":1}}`)
	}))
	defer srv.Close()

	p := NewAnthropicProvider("test-key", srv.URL)

	req := &llm.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: llm.StringContent("hi")},
		},
		Tools: []llm.Tool{
			{
				Type: "function",
				Function: llm.Function{
					Name:        "search",
					Description: "Web search",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`),
				},
			},
		},
		ToolChoice: json.RawMessage(`"required"`),
	}

	if _, err := p.ChatCompletion(t.Context(), req); err != nil {
		t.Fatalf("ChatCompletion error: %v", err)
	}

	// Verify tools were translated correctly.
	if len(captured.Tools) != 1 {
		t.Fatalf("captured tools count = %d, want 1", len(captured.Tools))
	}
	tool := captured.Tools[0]
	if tool.Name != "search" {
		t.Errorf("tool.name = %q, want search", tool.Name)
	}
	if tool.Description != "Web search" {
		t.Errorf("tool.description = %q, want 'Web search'", tool.Description)
	}
	if tool.InputSchema == nil {
		t.Error("tool.input_schema is nil")
	}

	// Verify tool_choice "required" → Anthropic "any".
	if captured.ToolChoice == nil {
		t.Fatal("tool_choice is nil")
	}
	if captured.ToolChoice.Type != "any" {
		t.Errorf("tool_choice.type = %q, want any", captured.ToolChoice.Type)
	}
}

// --------------------------------------------------------------------------
// Streaming tool-call round-trip
// --------------------------------------------------------------------------

func TestAnthropicProvider_ToolCall_Streaming(t *testing.T) {
	// Simulate Anthropic streaming a tool_use block.
	sseEvents := []string{
		`{"type":"message_start","message":{"id":"msg_stream_001","type":"message","role":"assistant","content":[],"model":"claude-3-5-sonnet-20241022","stop_reason":null,"usage":{"input_tokens":20,"output_tokens":0}}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_stream_01","name":"get_weather","input":{}}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"location\":"}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"London\"}"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":12}}`,
		`{"type":"message_stop"}`,
	}

	srv := fakeAnthropicSSEServer(t, sseEvents)
	defer srv.Close()

	p := NewAnthropicProvider("test-key", srv.URL)

	req := &llm.ChatCompletionRequest{
		Model:  "claude-3-5-sonnet-20241022",
		Stream: true,
		Messages: []llm.Message{
			{Role: "user", Content: llm.StringContent("Weather in London?")},
		},
		Tools: []llm.Tool{{
			Type:     "function",
			Function: llm.Function{Name: "get_weather"},
		}},
	}

	stream, err := p.ChatCompletionStream(t.Context(), req)
	if err != nil {
		t.Fatalf("ChatCompletionStream error: %v", err)
	}
	defer stream.Close()

	// Collect all SSE chunks from the translated stream.
	var chunks []llm.ChatCompletionChunk
	scanner := bufio.NewScanner(stream)
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
			t.Fatalf("unmarshal chunk: %v", err)
		}
		chunks = append(chunks, chunk)
	}

	if len(chunks) == 0 {
		t.Fatal("received no chunks")
	}

	// Find the content_block_start chunk (should have tool call id + name).
	var nameChunk *llm.ChatCompletionChunk
	for i := range chunks {
		c := &chunks[i]
		if len(c.Choices) > 0 && c.Choices[0].Delta != nil &&
			len(c.Choices[0].Delta.ToolCalls) > 0 &&
			c.Choices[0].Delta.ToolCalls[0].Function.Name != "" {
			nameChunk = c
			break
		}
	}
	if nameChunk == nil {
		t.Fatal("no chunk with tool call name found")
	}
	tcd := nameChunk.Choices[0].Delta.ToolCalls[0]
	if tcd.ID != "toolu_stream_01" {
		t.Errorf("tool_call.id = %q, want toolu_stream_01", tcd.ID)
	}
	if tcd.Type != "function" {
		t.Errorf("tool_call.type = %q, want function", tcd.Type)
	}
	if tcd.Function.Name != "get_weather" {
		t.Errorf("tool_call.function.name = %q, want get_weather", tcd.Function.Name)
	}

	// Collect all argument deltas and verify they form valid JSON when concatenated.
	var argsBuilder strings.Builder
	for _, c := range chunks {
		if len(c.Choices) == 0 || c.Choices[0].Delta == nil {
			continue
		}
		for _, tcd := range c.Choices[0].Delta.ToolCalls {
			argsBuilder.WriteString(tcd.Function.Arguments)
		}
	}
	fullArgs := argsBuilder.String()
	var parsedArgs map[string]string
	if err := json.Unmarshal([]byte(fullArgs), &parsedArgs); err != nil {
		t.Fatalf("accumulated arguments %q are not valid JSON: %v", fullArgs, err)
	}
	if parsedArgs["location"] != "London" {
		t.Errorf("arguments.location = %q, want London", parsedArgs["location"])
	}

	// Find the finish chunk with stop_reason = "tool_calls".
	var finishReason string
	for _, c := range chunks {
		if len(c.Choices) > 0 && c.Choices[0].FinishReason != "" {
			finishReason = c.Choices[0].FinishReason
		}
	}
	if finishReason != "tool_calls" {
		t.Errorf("finish_reason = %q, want tool_calls", finishReason)
	}
}
