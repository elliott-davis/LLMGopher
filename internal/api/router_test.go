package api_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ed007183/llmgopher/internal/api"
	"github.com/ed007183/llmgopher/internal/mocks"
	"github.com/ed007183/llmgopher/pkg/llm"
)

var discardLogger = slog.New(slog.NewTextHandler(
	&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError + 1},
))

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func newTestDeps(provider *mocks.MockProvider) (*api.Dependencies, *mocks.MockAuditLogger) {
	registry := llm.NewRegistry()
	if provider != nil {
		registry.Register(provider, "gpt-4*", "claude*", "test-model*")
	}

	audit := &mocks.MockAuditLogger{}

	return &api.Dependencies{
		Registry:      registry,
		RateLimiter:   &mocks.MockRateLimiter{Allowed: true},
		Guardrail:     &mocks.MockGuardrail{Verdict: &llm.GuardrailVerdict{Allowed: true}},
		AuditLogger:   audit,
		BudgetTracker: &mocks.MockBudgetTracker{Remaining: 1e9},
		Pricing:       mocks.NewMockPricingLookup(),
		APIKeys:       map[string]string{"sk-test": "key-001"},
		Logger:        discardLogger,
	}, audit
}

func authedRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer sk-test")
	req.Header.Set("Content-Type", "application/json")
	return req
}

func waitForAuditEntries(audit *mocks.MockAuditLogger, minCount int) []*llm.AuditEntry {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		entries := audit.GetEntries()
		if len(entries) >= minCount {
			return entries
		}
		time.Sleep(10 * time.Millisecond)
	}
	return audit.GetEntries()
}

// --------------------------------------------------------------------------
// Health / Ready
// --------------------------------------------------------------------------

func TestHealth(t *testing.T) {
	deps, _ := newTestDeps(nil)
	handler := api.NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
}

func TestReady(t *testing.T) {
	deps, _ := newTestDeps(nil)
	handler := api.NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --------------------------------------------------------------------------
// POST /v1/chat/completions — sync
// --------------------------------------------------------------------------

func TestChatCompletions_Sync_Success(t *testing.T) {
	provider := &mocks.MockProvider{
		ProviderName: "test-provider",
		ChatResponse: &llm.ChatCompletionResponse{
			ID:      "chatcmpl-test-1",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o",
			Choices: []llm.Choice{
				{
					Index:        0,
					Message:      &llm.Message{Role: "assistant", Content: llm.StringContent("Hello from mock!")},
					FinishReason: "stop",
				},
			},
			Usage: &llm.Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		},
	}

	deps, _ := newTestDeps(provider)
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp llm.ChatCompletionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.ID != "chatcmpl-test-1" {
		t.Errorf("id = %q, want %q", resp.ID, "chatcmpl-test-1")
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("choices length = %d, want 1", len(resp.Choices))
	}
	if resp.Choices[0].Message.ContentString() != "Hello from mock!" {
		t.Errorf("content = %q, want %q", resp.Choices[0].Message.ContentString(), "Hello from mock!")
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("total_tokens = %d, want 15", resp.Usage.TotalTokens)
	}

	if len(provider.SyncCalls) != 1 {
		t.Errorf("provider received %d sync calls, want 1", len(provider.SyncCalls))
	}
}

func TestChatCompletions_MissingModel_Returns400(t *testing.T) {
	deps, _ := newTestDeps(&mocks.MockProvider{ProviderName: "test"})
	handler := api.NewRouter(deps)

	body := `{"messages":[{"role":"user","content":"hi"}]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChatCompletions_EmptyMessages_Returns400(t *testing.T) {
	deps, _ := newTestDeps(&mocks.MockProvider{ProviderName: "test"})
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChatCompletions_InvalidJSON_Returns400(t *testing.T) {
	deps, _ := newTestDeps(&mocks.MockProvider{ProviderName: "test"})
	handler := api.NewRouter(deps)

	req := authedRequest(http.MethodPost, "/v1/chat/completions", "not json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChatCompletions_UnknownModel_Returns400(t *testing.T) {
	deps, _ := newTestDeps(&mocks.MockProvider{ProviderName: "test"})
	handler := api.NewRouter(deps)

	body := `{"model":"unknown-model","messages":[{"role":"user","content":"hi"}]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChatCompletions_ProviderError_Returns502(t *testing.T) {
	provider := &mocks.MockProvider{
		ProviderName: "test",
		Err:          fmt.Errorf("upstream timeout"),
	}

	deps, _ := newTestDeps(provider)
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// --------------------------------------------------------------------------
// POST /v1/completions — sync + streaming
// --------------------------------------------------------------------------

func TestCompletions_Sync_Success(t *testing.T) {
	provider := &mocks.MockProvider{
		ProviderName: "test-provider",
		ChatResponse: &llm.ChatCompletionResponse{
			ID:      "chatcmpl-test-completion-1",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o",
			Choices: []llm.Choice{
				{
					Index:        0,
					Message:      &llm.Message{Role: "assistant", Content: llm.StringContent("Hello from completion!")},
					FinishReason: "stop",
				},
			},
			Usage: &llm.Usage{
				PromptTokens:     4,
				CompletionTokens: 3,
				TotalTokens:      7,
			},
		},
	}

	deps, audit := newTestDeps(provider)
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","prompt":"Say hello"}`
	req := authedRequest(http.MethodPost, "/v1/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp llm.CompletionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Object != "text_completion" {
		t.Errorf("object = %q, want %q", resp.Object, "text_completion")
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("choices length = %d, want 1", len(resp.Choices))
	}
	if resp.Choices[0].Text != "Hello from completion!" {
		t.Errorf("text = %q, want %q", resp.Choices[0].Text, "Hello from completion!")
	}

	if len(provider.SyncCalls) != 1 {
		t.Fatalf("provider sync calls = %d, want 1", len(provider.SyncCalls))
	}
	if len(provider.SyncCalls[0].Messages) != 1 {
		t.Fatalf("translated messages = %d, want 1", len(provider.SyncCalls[0].Messages))
	}
	if provider.SyncCalls[0].Messages[0].ContentString() != "Say hello" {
		t.Errorf("translated prompt = %q, want %q", provider.SyncCalls[0].Messages[0].ContentString(), "Say hello")
	}

	entries := waitForAuditEntries(audit, 1)
	if len(entries) == 0 {
		t.Fatal("expected at least one audit entry")
	}
	last := entries[len(entries)-1]
	if last.PromptTokens != 4 || last.OutputTokens != 3 || last.TotalTokens != 7 {
		t.Errorf("audit tokens = (%d,%d,%d), want (4,3,7)", last.PromptTokens, last.OutputTokens, last.TotalTokens)
	}
	if last.StatusCode != http.StatusOK {
		t.Errorf("audit status = %d, want %d", last.StatusCode, http.StatusOK)
	}
}

func TestCompletions_PromptArray_Returns400(t *testing.T) {
	deps, _ := newTestDeps(&mocks.MockProvider{ProviderName: "test-provider"})
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","prompt":["first","second"]}`
	req := authedRequest(http.MethodPost, "/v1/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "prompt must be a string; prompt arrays are not supported") {
		t.Errorf("expected clear array-prompt error message, got: %s", w.Body.String())
	}
}

func TestCompletions_Stream_SSE(t *testing.T) {
	chunks := []llm.ChatCompletionChunk{
		{
			ID: "chatcmpl-stream-c-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{Role: "assistant"}}},
		},
		{
			ID: "chatcmpl-stream-c-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{Content: llm.StringContent("Hello")}}},
		},
		{
			ID: "chatcmpl-stream-c-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{Content: llm.StringContent(" world")}}},
		},
		{
			ID: "chatcmpl-stream-c-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{}, FinishReason: "stop"}},
			Usage:   &llm.Usage{CompletionTokens: 2},
		},
	}

	provider := &mocks.MockProvider{
		ProviderName: "test-provider",
		StreamBody:   buildSSEStream(chunks),
	}

	deps, _ := newTestDeps(provider)
	handler := api.NewRouter(deps)
	body := `{"model":"gpt-4o","prompt":"hello","stream":true}`

	srv := httptest.NewServer(handler)
	defer srv.Close()

	httpReq, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/completions", strings.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	httpReq.Header.Set("Authorization", "Bearer sk-test")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d; body = %s", resp.StatusCode, http.StatusOK, string(respBody))
	}

	var received []llm.CompletionResponse
	var gotDone bool
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			gotDone = true
			break
		}

		var chunk llm.CompletionResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			t.Fatalf("unmarshal completions chunk: %v (data=%s)", err, data)
		}
		received = append(received, chunk)
	}

	if !gotDone {
		t.Fatal("did not receive [DONE] sentinel")
	}
	if len(received) != len(chunks) {
		t.Fatalf("received %d chunks, want %d", len(received), len(chunks))
	}
	for i, chunk := range received {
		if chunk.Object != "text_completion" {
			t.Fatalf("chunk[%d].object = %q, want text_completion", i, chunk.Object)
		}
	}

	var fullText strings.Builder
	for _, c := range received {
		if len(c.Choices) > 0 {
			fullText.WriteString(c.Choices[0].Text)
		}
	}
	if fullText.String() != "Hello world" {
		t.Errorf("accumulated text = %q, want %q", fullText.String(), "Hello world")
	}

	last := received[len(received)-1]
	if last.Choices[0].FinishReason != "stop" {
		t.Errorf("finish_reason = %q, want %q", last.Choices[0].FinishReason, "stop")
	}
	if last.Usage == nil || last.Usage.CompletionTokens != 2 {
		t.Errorf("usage = %+v, want CompletionTokens=2", last.Usage)
	}

	if len(provider.StreamCalls) != 1 {
		t.Fatalf("provider stream calls = %d, want 1", len(provider.StreamCalls))
	}
	if len(provider.StreamCalls[0].Messages) != 1 {
		t.Fatalf("translated stream messages = %d, want 1", len(provider.StreamCalls[0].Messages))
	}
	if provider.StreamCalls[0].Messages[0].ContentString() != "hello" {
		t.Errorf("translated stream prompt = %q, want %q", provider.StreamCalls[0].Messages[0].ContentString(), "hello")
	}
}

func TestChatCompletions_NoAuth_Returns401(t *testing.T) {
	deps, _ := newTestDeps(&mocks.MockProvider{ProviderName: "test"})
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestChatCompletions_RateLimited_Returns429(t *testing.T) {
	provider := &mocks.MockProvider{ProviderName: "test"}
	deps, _ := newTestDeps(provider)
	deps.RateLimiter = &mocks.MockRateLimiter{Allowed: false}
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestChatCompletions_GuardrailBlocked_Returns400(t *testing.T) {
	provider := &mocks.MockProvider{ProviderName: "test"}
	deps, _ := newTestDeps(provider)
	deps.Guardrail = &mocks.MockGuardrail{
		Verdict: &llm.GuardrailVerdict{Allowed: false, Reason: "policy violation"},
	}
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"bad content"}]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "policy violation") {
		t.Errorf("body should contain reason, got: %s", w.Body.String())
	}
}

// --------------------------------------------------------------------------
// POST /v1/chat/completions — streaming SSE
// --------------------------------------------------------------------------

func buildSSEStream(chunks []llm.ChatCompletionChunk) io.ReadCloser {
	var buf strings.Builder
	for _, chunk := range chunks {
		data, _ := json.Marshal(chunk)
		buf.WriteString("data: ")
		buf.Write(data)
		buf.WriteString("\n\n")
	}
	buf.WriteString("data: [DONE]\n\n")
	return io.NopCloser(strings.NewReader(buf.String()))
}

func TestChatCompletions_Stream_SSE(t *testing.T) {
	chunks := []llm.ChatCompletionChunk{
		{
			ID: "chatcmpl-stream-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{Role: "assistant"}}},
		},
		{
			ID: "chatcmpl-stream-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{Content: llm.StringContent("Hello")}}},
		},
		{
			ID: "chatcmpl-stream-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{Content: llm.StringContent(" world")}}},
		},
		{
			ID: "chatcmpl-stream-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.MessageDelta{}, FinishReason: "stop"}},
			Usage:   &llm.Usage{CompletionTokens: 2},
		},
	}

	provider := &mocks.MockProvider{
		ProviderName: "test-provider",
		StreamBody:   buildSSEStream(chunks),
	}

	deps, _ := newTestDeps(provider)
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stream":true}`

	// Use an httptest.Server so we get a real TCP connection with flushing support.
	srv := httptest.NewServer(handler)
	defer srv.Close()

	httpReq, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/chat/completions", strings.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	httpReq.Header.Set("Authorization", "Bearer sk-test")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d; body = %s", resp.StatusCode, http.StatusOK, string(respBody))
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	// Parse SSE events from the response.
	var receivedChunks []llm.ChatCompletionChunk
	var gotDone bool
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			gotDone = true
			break
		}
		var chunk llm.ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			t.Logf("skipping unparseable chunk: %s", data)
			continue
		}
		receivedChunks = append(receivedChunks, chunk)
	}

	if !gotDone {
		t.Error("did not receive [DONE] sentinel")
	}

	if len(receivedChunks) != len(chunks) {
		t.Fatalf("received %d chunks, want %d", len(receivedChunks), len(chunks))
	}

	// Verify the text deltas were forwarded correctly.
	var fullText strings.Builder
	for _, c := range receivedChunks {
		if len(c.Choices) > 0 && c.Choices[0].Delta != nil {
			fullText.WriteString(c.Choices[0].Delta.ContentString())
		}
	}
	if fullText.String() != "Hello world" {
		t.Errorf("accumulated text = %q, want %q", fullText.String(), "Hello world")
	}

	// Verify the last chunk has finish_reason and usage.
	last := receivedChunks[len(receivedChunks)-1]
	if last.Choices[0].FinishReason != "stop" {
		t.Errorf("last chunk finish_reason = %q, want %q", last.Choices[0].FinishReason, "stop")
	}
	if last.Usage == nil || last.Usage.CompletionTokens != 2 {
		t.Errorf("last chunk usage = %+v, want CompletionTokens=2", last.Usage)
	}

	if len(provider.StreamCalls) != 1 {
		t.Errorf("provider received %d stream calls, want 1", len(provider.StreamCalls))
	}
}

func TestChatCompletions_Stream_ProviderError_Returns502(t *testing.T) {
	provider := &mocks.MockProvider{
		ProviderName: "test",
		Err:          fmt.Errorf("stream init failed"),
	}

	deps, _ := newTestDeps(provider)
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stream":true}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// --------------------------------------------------------------------------
// Tool calling — gateway round-trip
// --------------------------------------------------------------------------

// TestChatCompletions_ToolCall_RequestPassthrough verifies that tools and
// tool_choice in the request body are forwarded to the provider unchanged.
func TestChatCompletions_ToolCall_RequestPassthrough(t *testing.T) {
	provider := &mocks.MockProvider{
		ProviderName: "test-provider",
		ChatResponse: &llm.ChatCompletionResponse{
			ID: "chatcmpl-tc-1", Object: "chat.completion",
			Created: time.Now().Unix(), Model: "gpt-4o",
			Choices: []llm.Choice{{
				Index:        0,
				Message:      &llm.Message{Role: "assistant", Content: llm.StringContent("")},
				FinishReason: "tool_calls",
			}},
			Usage: &llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
		},
	}

	deps, _ := newTestDeps(provider)
	handler := api.NewRouter(deps)

	body := `{
		"model": "gpt-4o",
		"messages": [{"role":"user","content":"What is the weather?"}],
		"tools": [{"type":"function","function":{"name":"get_weather","description":"Get weather","parameters":{"type":"object","properties":{"location":{"type":"string"}},"required":["location"]}}}],
		"tool_choice": "auto"
	}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	if len(provider.SyncCalls) != 1 {
		t.Fatalf("provider call count = %d, want 1", len(provider.SyncCalls))
	}

	got := provider.SyncCalls[0]

	// Tools were passed through to the provider.
	if len(got.Tools) != 1 {
		t.Fatalf("tools count = %d, want 1", len(got.Tools))
	}
	if got.Tools[0].Function.Name != "get_weather" {
		t.Errorf("tool name = %q, want get_weather", got.Tools[0].Function.Name)
	}

	// tool_choice was passed through to the provider.
	if got.ToolChoice == nil {
		t.Fatal("tool_choice is nil")
	}
	var tc string
	if err := json.Unmarshal(got.ToolChoice, &tc); err != nil || tc != "auto" {
		t.Errorf("tool_choice = %q, want auto", tc)
	}
}

// TestChatCompletions_ToolCall_ResponseRoundTrip verifies that a provider
// response containing tool_calls is serialized back to the client correctly.
func TestChatCompletions_ToolCall_ResponseRoundTrip(t *testing.T) {
	provider := &mocks.MockProvider{
		ProviderName: "test-provider",
		ChatResponse: &llm.ChatCompletionResponse{
			ID: "chatcmpl-tc-2", Object: "chat.completion",
			Created: time.Now().Unix(), Model: "gpt-4o",
			Choices: []llm.Choice{{
				Index: 0,
				Message: &llm.Message{
					Role: "assistant",
					ToolCalls: []llm.ToolCall{{
						ID:   "call_abc123",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "get_weather",
							Arguments: `{"location":"Seattle","unit":"celsius"}`,
						},
					}},
				},
				FinishReason: "tool_calls",
			}},
			Usage: &llm.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
		},
	}

	deps, _ := newTestDeps(provider)
	handler := api.NewRouter(deps)

	body := `{"model":"gpt-4o","messages":[{"role":"user","content":"Weather in Seattle?"}],"tools":[{"type":"function","function":{"name":"get_weather","parameters":{}}}]}`
	req := authedRequest(http.MethodPost, "/v1/chat/completions", body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp llm.ChatCompletionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("choices count = %d, want 1", len(resp.Choices))
	}
	msg := resp.Choices[0].Message
	if msg == nil {
		t.Fatal("message is nil")
	}
	if resp.Choices[0].FinishReason != "tool_calls" {
		t.Errorf("finish_reason = %q, want tool_calls", resp.Choices[0].FinishReason)
	}
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("tool_calls count = %d, want 1", len(msg.ToolCalls))
	}
	tc := msg.ToolCalls[0]
	if tc.ID != "call_abc123" {
		t.Errorf("tool_call.id = %q, want call_abc123", tc.ID)
	}
	if tc.Type != "function" {
		t.Errorf("tool_call.type = %q, want function", tc.Type)
	}
	if tc.Function.Name != "get_weather" {
		t.Errorf("tool_call.function.name = %q, want get_weather", tc.Function.Name)
	}

	// Arguments should be valid JSON with the expected content.
	var args map[string]string
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		t.Fatalf("unmarshal arguments: %v", err)
	}
	if args["location"] != "Seattle" {
		t.Errorf("arguments.location = %q, want Seattle", args["location"])
	}
}
