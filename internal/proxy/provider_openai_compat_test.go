package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestOpenAICompatProviderChatCompletion_UsesCompatPathAndAuth(t *testing.T) {
	t.Parallel()

	var (
		gotPath          string
		gotAuth          string
		gotStreamRequest bool
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if stream, ok := payload["stream"].(bool); ok {
			gotStreamRequest = stream
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cmpl_1","object":"chat.completion","created":1,"model":"llama3","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatProvider("groq", server.URL+"/v1", "groq-key")
	resp, err := provider.ChatCompletion(t.Context(), &llm.ChatCompletionRequest{
		Model:    "llama3",
		Stream:   true, // Should be forced to false for sync calls.
		Messages: []llm.Message{{Role: "user", Content: llm.StringContent("hello")}},
	})
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v, want nil", err)
	}

	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q, want %q", gotPath, "/v1/chat/completions")
	}
	if gotAuth != "Bearer groq-key" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer groq-key")
	}
	if gotStreamRequest {
		t.Fatalf("request stream = %t, want false", gotStreamRequest)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message == nil || resp.Choices[0].Message.ContentString() != "ok" {
		t.Fatalf("unexpected response payload: %+v", resp)
	}
}

func TestOpenAICompatProviderChatCompletion_OmitsAuthWhenKeyMissing(t *testing.T) {
	t.Parallel()

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cmpl_2","object":"chat.completion","created":1,"model":"llama3","choices":[{"index":0,"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatProvider("ollama", server.URL+"/v1", "")
	_, err := provider.ChatCompletion(t.Context(), &llm.ChatCompletionRequest{
		Model:    "llama3",
		Messages: []llm.Message{{Role: "user", Content: llm.StringContent("hello")}},
	})
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v, want nil", err)
	}

	if gotAuth != "" {
		t.Fatalf("Authorization = %q, want empty", gotAuth)
	}
}

func TestResolveProviderBaseURL_WellKnownFallback(t *testing.T) {
	t.Parallel()

	if got := ResolveProviderBaseURL("Groq", ""); got != WellKnownBaseURLs["groq"] {
		t.Fatalf("ResolveProviderBaseURL() = %q, want %q", got, WellKnownBaseURLs["groq"])
	}
}
