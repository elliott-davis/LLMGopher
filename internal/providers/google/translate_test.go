package google

import (
	"encoding/json"
	"testing"

	vertex "cloud.google.com/go/vertexai/genai"
	gemini "github.com/google/generative-ai-go/genai"

	"github.com/ed007183/llmgopher/pkg/llm"
)

func float64Ptr(v float64) *float64 { return &v }
func intPtr(v int) *int             { return &v }

// --- parseStopSequences ---

func TestParseStopSequences_Array(t *testing.T) {
	raw := json.RawMessage(`["stop1","stop2"]`)
	got := parseStopSequences(raw)
	if len(got) != 2 || got[0] != "stop1" || got[1] != "stop2" {
		t.Errorf("parseStopSequences array = %v, want [stop1 stop2]", got)
	}
}

func TestParseStopSequences_String(t *testing.T) {
	raw := json.RawMessage(`"halt"`)
	got := parseStopSequences(raw)
	if len(got) != 1 || got[0] != "halt" {
		t.Errorf("parseStopSequences string = %v, want [halt]", got)
	}
}

func TestParseStopSequences_Nil(t *testing.T) {
	got := parseStopSequences(nil)
	if len(got) != 0 {
		t.Errorf("parseStopSequences nil = %v, want empty", got)
	}
}

// --- stripPrefix ---

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		model, prefix, want string
	}{
		{"gemini/gemini-1.5-pro", "gemini/", "gemini-1.5-pro"},
		{"vertex_ai/gemini-1.5-flash", "vertex_ai/", "gemini-1.5-flash"},
		{"gpt-4o", "gemini/", "gpt-4o"},
		{"gemini/", "gemini/", ""},
	}
	for _, tt := range tests {
		got := stripPrefix(tt.model, tt.prefix)
		if got != tt.want {
			t.Errorf("stripPrefix(%q, %q) = %q, want %q", tt.model, tt.prefix, got, tt.want)
		}
	}
}

// --- buildGeminiModel ---

func TestBuildGeminiModel_SystemInstruction(t *testing.T) {
	// We can't call client.GenerativeModel without a real client, so we test
	// the message partitioning logic indirectly by checking that the helper
	// correctly separates system messages from history.
	req := &llm.ChatCompletionRequest{
		Model: "gemini/gemini-1.5-pro",
		Messages: []llm.Message{
			{Role: "system", Content: llm.StringContent("You are a helpful assistant.")},
			{Role: "user", Content: llm.StringContent("Hello")},
			{Role: "assistant", Content: llm.StringContent("Hi there!")},
			{Role: "user", Content: llm.StringContent("How are you?")},
		},
		Temperature: float64Ptr(0.7),
		MaxTokens:   intPtr(1024),
	}

	// We need a real client to test buildGeminiModel, but we can test the
	// partitioning logic by examining what gets set on the model.
	// Since we can't create a real client in unit tests, we'll test the
	// translation output functions instead and verify the provider compiles.
	_ = req
}

func TestGeminiContentPartsFromMessage_URLImage(t *testing.T) {
	msg := llm.Message{
		Role: "user",
		Content: json.RawMessage(
			`[{"type":"text","text":"describe"},{"type":"image_url","image_url":{"url":"https://example.com/cat.png"}}]`,
		),
	}

	parts := geminiContentPartsFromMessage(msg)
	if len(parts) != 2 {
		t.Fatalf("len(parts) = %d, want 2", len(parts))
	}
	textPart, ok := parts[0].(gemini.Text)
	if !ok || string(textPart) != "describe" {
		t.Fatalf("parts[0] = %T(%v), want Text(describe)", parts[0], parts[0])
	}
	filePart, ok := parts[1].(gemini.FileData)
	if !ok {
		t.Fatalf("parts[1] = %T, want FileData", parts[1])
	}
	if filePart.URI != "https://example.com/cat.png" {
		t.Errorf("file URI = %q, want https://example.com/cat.png", filePart.URI)
	}
	if filePart.MIMEType != "image/png" {
		t.Errorf("file MIMEType = %q, want image/png", filePart.MIMEType)
	}
}

func TestGeminiContentPartsFromMessage_DataURIImage(t *testing.T) {
	msg := llm.Message{
		Role:    "user",
		Content: json.RawMessage(`[{"type":"image_url","image_url":{"url":"data:image/png;base64,aGVsbG8="}}]`),
	}

	parts := geminiContentPartsFromMessage(msg)
	if len(parts) != 1 {
		t.Fatalf("len(parts) = %d, want 1", len(parts))
	}
	blob, ok := parts[0].(gemini.Blob)
	if !ok {
		t.Fatalf("parts[0] = %T, want Blob", parts[0])
	}
	if blob.MIMEType != "image/png" {
		t.Errorf("blob MIMEType = %q, want image/png", blob.MIMEType)
	}
	if string(blob.Data) != "hello" {
		t.Errorf("blob Data = %q, want hello", string(blob.Data))
	}
}

func TestVertexContentPartsFromMessage_URLAndDataURIImages(t *testing.T) {
	msg := llm.Message{
		Role: "user",
		Content: json.RawMessage(
			`[{"type":"image_url","image_url":{"url":"https://example.com/cat.webp"}},{"type":"image_url","image_url":{"url":"data:image/jpeg;base64,aGVsbG8="}}]`,
		),
	}

	parts := vertexContentPartsFromMessage(msg)
	if len(parts) != 2 {
		t.Fatalf("len(parts) = %d, want 2", len(parts))
	}
	filePart, ok := parts[0].(vertex.FileData)
	if !ok {
		t.Fatalf("parts[0] = %T, want FileData", parts[0])
	}
	if filePart.FileURI != "https://example.com/cat.webp" {
		t.Errorf("file FileURI = %q, want https://example.com/cat.webp", filePart.FileURI)
	}
	if filePart.MIMEType != "image/webp" {
		t.Errorf("file MIMEType = %q, want image/webp", filePart.MIMEType)
	}
	blob, ok := parts[1].(vertex.Blob)
	if !ok {
		t.Fatalf("parts[1] = %T, want Blob", parts[1])
	}
	if blob.MIMEType != "image/jpeg" {
		t.Errorf("blob MIMEType = %q, want image/jpeg", blob.MIMEType)
	}
	if string(blob.Data) != "hello" {
		t.Errorf("blob Data = %q, want hello", string(blob.Data))
	}
}

// --- Gemini response translation ---

func TestGeminiResponseToOpenAI_Basic(t *testing.T) {
	resp := &gemini.GenerateContentResponse{
		Candidates: []*gemini.Candidate{
			{
				Content: &gemini.Content{
					Role:  "model",
					Parts: []gemini.Part{gemini.Text("Hello, world!")},
				},
				FinishReason: gemini.FinishReasonStop,
			},
		},
		UsageMetadata: &gemini.UsageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 5,
			TotalTokenCount:      15,
		},
	}

	out := geminiResponseToOpenAI(resp, "gemini-1.5-pro", "chatcmpl-abc123", 1700000000)

	if out.ID != "chatcmpl-abc123" {
		t.Errorf("ID = %q, want chatcmpl-abc123", out.ID)
	}
	if out.Object != "chat.completion" {
		t.Errorf("Object = %q, want chat.completion", out.Object)
	}
	if out.Model != "gemini-1.5-pro" {
		t.Errorf("Model = %q, want gemini-1.5-pro", out.Model)
	}
	if len(out.Choices) != 1 {
		t.Fatalf("len(Choices) = %d, want 1", len(out.Choices))
	}
	if out.Choices[0].Message.ContentString() != "Hello, world!" {
		t.Errorf("Content = %q, want Hello, world!", out.Choices[0].Message.ContentString())
	}
	if out.Choices[0].Message.Role != "assistant" {
		t.Errorf("Role = %q, want assistant", out.Choices[0].Message.Role)
	}
	if out.Choices[0].FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want stop", out.Choices[0].FinishReason)
	}
	if out.Usage == nil {
		t.Fatal("Usage is nil")
	}
	if out.Usage.PromptTokens != 10 {
		t.Errorf("PromptTokens = %d, want 10", out.Usage.PromptTokens)
	}
	if out.Usage.CompletionTokens != 5 {
		t.Errorf("CompletionTokens = %d, want 5", out.Usage.CompletionTokens)
	}
	if out.Usage.TotalTokens != 15 {
		t.Errorf("TotalTokens = %d, want 15", out.Usage.TotalTokens)
	}
}

func TestGeminiResponseToOpenAI_NilContent(t *testing.T) {
	resp := &gemini.GenerateContentResponse{
		Candidates: []*gemini.Candidate{
			{
				Content:      nil,
				FinishReason: gemini.FinishReasonSafety,
			},
		},
	}

	out := geminiResponseToOpenAI(resp, "gemini-1.5-pro", "chatcmpl-xyz", 1700000000)

	if len(out.Choices) != 1 {
		t.Fatalf("len(Choices) = %d, want 1", len(out.Choices))
	}
	if out.Choices[0].Message.ContentString() != "" {
		t.Errorf("Content = %q, want empty", out.Choices[0].Message.ContentString())
	}
	if out.Choices[0].FinishReason != "content_filter" {
		t.Errorf("FinishReason = %q, want content_filter", out.Choices[0].FinishReason)
	}
	if out.Usage != nil {
		t.Errorf("Usage = %v, want nil", out.Usage)
	}
}

func TestGeminiResponseToOpenAI_MultiPart(t *testing.T) {
	resp := &gemini.GenerateContentResponse{
		Candidates: []*gemini.Candidate{
			{
				Content: &gemini.Content{
					Role: "model",
					Parts: []gemini.Part{
						gemini.Text("Part one. "),
						gemini.Text("Part two."),
					},
				},
				FinishReason: gemini.FinishReasonStop,
			},
		},
	}

	out := geminiResponseToOpenAI(resp, "gemini-1.5-pro", "chatcmpl-multi", 1700000000)
	if out.Choices[0].Message.ContentString() != "Part one. Part two." {
		t.Errorf("Content = %q, want 'Part one. Part two.'", out.Choices[0].Message.ContentString())
	}
}

func TestGeminiResponseToOpenAI_FunctionCall(t *testing.T) {
	resp := &gemini.GenerateContentResponse{
		Candidates: []*gemini.Candidate{
			{
				Content: &gemini.Content{
					Role: "model",
					Parts: []gemini.Part{
						gemini.FunctionCall{
							Name: "get_weather",
							Args: map[string]any{"location": "Seattle"},
						},
					},
				},
				FinishReason: gemini.FinishReasonStop,
			},
		},
	}

	out := geminiResponseToOpenAI(resp, "gemini-1.5-pro", "chatcmpl-fc", 1700000000)

	if len(out.Choices) != 1 {
		t.Fatalf("len(Choices) = %d, want 1", len(out.Choices))
	}
	if out.Choices[0].FinishReason != "tool_calls" {
		t.Errorf("FinishReason = %q, want tool_calls", out.Choices[0].FinishReason)
	}
	msg := out.Choices[0].Message
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("tool_calls count = %d, want 1", len(msg.ToolCalls))
	}
	if msg.ToolCalls[0].Function.Name != "get_weather" {
		t.Errorf("tool_call.function.name = %q, want get_weather", msg.ToolCalls[0].Function.Name)
	}
}

// --- Gemini finish reason mapping ---

func TestMapGeminiFinishReason(t *testing.T) {
	tests := []struct {
		in   gemini.FinishReason
		want string
	}{
		{gemini.FinishReasonStop, "stop"},
		{gemini.FinishReasonMaxTokens, "length"},
		{gemini.FinishReasonSafety, "content_filter"},
		{gemini.FinishReasonRecitation, "content_filter"},
		{gemini.FinishReasonOther, "stop"},
		{gemini.FinishReasonUnspecified, ""},
	}
	for _, tt := range tests {
		got := mapGeminiFinishReason(tt.in)
		if got != tt.want {
			t.Errorf("mapGeminiFinishReason(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// --- jsonSchemaToGeminiSchema ---

func TestJSONSchemaToGeminiSchema_Object(t *testing.T) {
	raw := json.RawMessage(`{
		"type": "object",
		"description": "Weather parameters",
		"properties": {
			"location": {"type": "string", "description": "City name"},
			"unit": {"type": "string", "enum": ["celsius","fahrenheit"]}
		},
		"required": ["location"]
	}`)

	schema := jsonSchemaToGeminiSchema(raw)
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if schema.Type != gemini.TypeObject {
		t.Errorf("type = %v, want TypeObject", schema.Type)
	}
	if schema.Description != "Weather parameters" {
		t.Errorf("description = %q, want 'Weather parameters'", schema.Description)
	}
	if len(schema.Properties) != 2 {
		t.Errorf("properties count = %d, want 2", len(schema.Properties))
	}
	if schema.Properties["location"].Type != gemini.TypeString {
		t.Errorf("location type = %v, want TypeString", schema.Properties["location"].Type)
	}
	unitSchema := schema.Properties["unit"]
	if len(unitSchema.Enum) != 2 {
		t.Errorf("unit enum count = %d, want 2", len(unitSchema.Enum))
	}
	if len(schema.Required) != 1 || schema.Required[0] != "location" {
		t.Errorf("required = %v, want [location]", schema.Required)
	}
}

func TestJSONSchemaToGeminiSchema_Nil(t *testing.T) {
	schema := jsonSchemaToGeminiSchema(nil)
	if schema != nil {
		t.Errorf("expected nil for nil input, got %+v", schema)
	}
}

// --- buildGeminiToolConfig ---

func TestBuildGeminiToolConfig_StringForms(t *testing.T) {
	tests := []struct {
		input    string
		wantMode gemini.FunctionCallingMode
	}{
		{`"none"`, gemini.FunctionCallingNone},
		{`"auto"`, gemini.FunctionCallingAuto},
		{`"required"`, gemini.FunctionCallingAny},
	}
	for _, tt := range tests {
		cfg := buildGeminiToolConfig(json.RawMessage(tt.input))
		if cfg == nil || cfg.FunctionCallingConfig == nil {
			t.Errorf("%s: config is nil", tt.input)
			continue
		}
		if cfg.FunctionCallingConfig.Mode != tt.wantMode {
			t.Errorf("%s: mode = %v, want %v", tt.input, cfg.FunctionCallingConfig.Mode, tt.wantMode)
		}
	}
}

func TestBuildGeminiToolConfig_FunctionObject(t *testing.T) {
	raw := json.RawMessage(`{"type":"function","function":{"name":"get_weather"}}`)
	cfg := buildGeminiToolConfig(raw)
	if cfg == nil || cfg.FunctionCallingConfig == nil {
		t.Fatal("config is nil")
	}
	if cfg.FunctionCallingConfig.Mode != gemini.FunctionCallingAny {
		t.Errorf("mode = %v, want FunctionCallingAny", cfg.FunctionCallingConfig.Mode)
	}
	if len(cfg.FunctionCallingConfig.AllowedFunctionNames) != 1 ||
		cfg.FunctionCallingConfig.AllowedFunctionNames[0] != "get_weather" {
		t.Errorf("allowed_function_names = %v, want [get_weather]", cfg.FunctionCallingConfig.AllowedFunctionNames)
	}
}

// --- Vertex response translation ---

func TestVertexResponseToOpenAI_Basic(t *testing.T) {
	resp := &vertex.GenerateContentResponse{
		Candidates: []*vertex.Candidate{
			{
				Content: &vertex.Content{
					Role:  "model",
					Parts: []vertex.Part{vertex.Text("Hello from Vertex!")},
				},
				FinishReason: vertex.FinishReasonStop,
			},
		},
		UsageMetadata: &vertex.UsageMetadata{
			PromptTokenCount:     20,
			CandidatesTokenCount: 8,
			TotalTokenCount:      28,
		},
	}

	out := vertexResponseToOpenAI(resp, "gemini-1.5-flash", "chatcmpl-vtx123", 1700000000)

	if out.ID != "chatcmpl-vtx123" {
		t.Errorf("ID = %q, want chatcmpl-vtx123", out.ID)
	}
	if out.Model != "gemini-1.5-flash" {
		t.Errorf("Model = %q, want gemini-1.5-flash", out.Model)
	}
	if len(out.Choices) != 1 {
		t.Fatalf("len(Choices) = %d, want 1", len(out.Choices))
	}
	if out.Choices[0].Message.ContentString() != "Hello from Vertex!" {
		t.Errorf("Content = %q, want Hello from Vertex!", out.Choices[0].Message.ContentString())
	}
	if out.Usage.PromptTokens != 20 || out.Usage.CompletionTokens != 8 || out.Usage.TotalTokens != 28 {
		t.Errorf("Usage = %+v, want {20, 8, 28}", out.Usage)
	}
}

func TestVertexResponseToOpenAI_NilContent(t *testing.T) {
	resp := &vertex.GenerateContentResponse{
		Candidates: []*vertex.Candidate{
			{
				Content:      nil,
				FinishReason: vertex.FinishReasonSafety,
			},
		},
	}

	out := vertexResponseToOpenAI(resp, "gemini-1.5-flash", "chatcmpl-vtx-nil", 1700000000)
	if out.Choices[0].Message.ContentString() != "" {
		t.Errorf("Content = %q, want empty", out.Choices[0].Message.ContentString())
	}
	if out.Choices[0].FinishReason != "content_filter" {
		t.Errorf("FinishReason = %q, want content_filter", out.Choices[0].FinishReason)
	}
}

func TestVertexResponseToOpenAI_FunctionCall(t *testing.T) {
	resp := &vertex.GenerateContentResponse{
		Candidates: []*vertex.Candidate{
			{
				Content: &vertex.Content{
					Role: "model",
					Parts: []vertex.Part{
						vertex.FunctionCall{
							Name: "get_weather",
							Args: map[string]any{"location": "Seattle"},
						},
					},
				},
				FinishReason: vertex.FinishReasonStop,
			},
		},
	}

	out := vertexResponseToOpenAI(resp, "gemini-1.5-pro", "chatcmpl-vtx-fc", 1700000000)

	if len(out.Choices) != 1 {
		t.Fatalf("len(Choices) = %d, want 1", len(out.Choices))
	}
	if out.Choices[0].FinishReason != "tool_calls" {
		t.Errorf("FinishReason = %q, want tool_calls", out.Choices[0].FinishReason)
	}
	msg := out.Choices[0].Message
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("tool_calls count = %d, want 1", len(msg.ToolCalls))
	}
	if msg.ToolCalls[0].Function.Name != "get_weather" {
		t.Errorf("tool_call.function.name = %q, want get_weather", msg.ToolCalls[0].Function.Name)
	}
}

// --- Vertex finish reason mapping ---

func TestMapVertexFinishReason(t *testing.T) {
	tests := []struct {
		in   vertex.FinishReason
		want string
	}{
		{vertex.FinishReasonStop, "stop"},
		{vertex.FinishReasonMaxTokens, "length"},
		{vertex.FinishReasonSafety, "content_filter"},
		{vertex.FinishReasonRecitation, "content_filter"},
		{vertex.FinishReasonOther, "stop"},
		{vertex.FinishReasonUnspecified, ""},
	}
	for _, tt := range tests {
		got := mapVertexFinishReason(tt.in)
		if got != tt.want {
			t.Errorf("mapVertexFinishReason(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// --- splitGeminiHistory ---

func TestSplitGeminiHistory_MultiTurn(t *testing.T) {
	contents := []*gemini.Content{
		{Role: "user", Parts: []gemini.Part{gemini.Text("Hello")}},
		{Role: "model", Parts: []gemini.Part{gemini.Text("Hi there")}},
		{Role: "user", Parts: []gemini.Part{gemini.Text("How are you?")}},
	}

	history, lastParts := splitGeminiHistory(contents)

	if len(history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(history))
	}
	if len(lastParts) != 1 {
		t.Fatalf("len(lastParts) = %d, want 1", len(lastParts))
	}
	if text, ok := lastParts[0].(gemini.Text); !ok || string(text) != "How are you?" {
		t.Errorf("lastParts[0] = %v, want Text('How are you?')", lastParts[0])
	}
}

func TestSplitGeminiHistory_SingleMessage(t *testing.T) {
	contents := []*gemini.Content{
		{Role: "user", Parts: []gemini.Part{gemini.Text("Hello")}},
	}

	history, lastParts := splitGeminiHistory(contents)

	if len(history) != 0 {
		t.Fatalf("len(history) = %d, want 0", len(history))
	}
	if len(lastParts) != 1 {
		t.Fatalf("len(lastParts) = %d, want 1", len(lastParts))
	}
}

func TestSplitGeminiHistory_Empty(t *testing.T) {
	history, lastParts := splitGeminiHistory(nil)

	if history != nil {
		t.Errorf("history = %v, want nil", history)
	}
	if len(lastParts) != 1 {
		t.Fatalf("len(lastParts) = %d, want 1 (empty text fallback)", len(lastParts))
	}
}

// --- splitVertexHistory ---

func TestSplitVertexHistory_MultiTurn(t *testing.T) {
	contents := []*vertex.Content{
		{Role: "user", Parts: []vertex.Part{vertex.Text("Hello")}},
		{Role: "model", Parts: []vertex.Part{vertex.Text("Hi there")}},
		{Role: "user", Parts: []vertex.Part{vertex.Text("How are you?")}},
	}

	history, lastParts := splitVertexHistory(contents)

	if len(history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(history))
	}
	if len(lastParts) != 1 {
		t.Fatalf("len(lastParts) = %d, want 1", len(lastParts))
	}
	if text, ok := lastParts[0].(vertex.Text); !ok || string(text) != "How are you?" {
		t.Errorf("lastParts[0] = %v, want Text('How are you?')", lastParts[0])
	}
}

func TestSplitVertexHistory_Empty(t *testing.T) {
	history, lastParts := splitVertexHistory(nil)

	if history != nil {
		t.Errorf("history = %v, want nil", history)
	}
	if len(lastParts) != 1 {
		t.Fatalf("len(lastParts) = %d, want 1 (empty text fallback)", len(lastParts))
	}
}

// --- Interface compliance ---

var _ llm.Provider = (*GeminiProvider)(nil)
var _ llm.Provider = (*VertexProvider)(nil)
var _ llm.EmbeddingProvider = (*GeminiProvider)(nil)
