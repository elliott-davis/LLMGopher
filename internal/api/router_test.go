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
					Message:      &llm.Message{Role: "assistant", Content: "Hello from mock!"},
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
	if resp.Choices[0].Message.Content != "Hello from mock!" {
		t.Errorf("content = %q, want %q", resp.Choices[0].Message.Content, "Hello from mock!")
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
			Choices: []llm.Choice{{Index: 0, Delta: &llm.Message{Role: "assistant"}}},
		},
		{
			ID: "chatcmpl-stream-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.Message{Content: "Hello"}}},
		},
		{
			ID: "chatcmpl-stream-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.Message{Content: " world"}}},
		},
		{
			ID: "chatcmpl-stream-1", Object: "chat.completion.chunk",
			Created: 1700000000, Model: "gpt-4o",
			Choices: []llm.Choice{{Index: 0, Delta: &llm.Message{}, FinishReason: "stop"}},
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
			fullText.WriteString(c.Choices[0].Delta.Content)
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
