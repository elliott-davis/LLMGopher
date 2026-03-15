package llm

import (
	"context"
	"io"
	"testing"
)

// stubProvider is a minimal Provider for registry tests.
type stubProvider struct{ name string }

func (s *stubProvider) Name() string { return s.name }
func (s *stubProvider) ChatCompletion(_ context.Context, _ *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return nil, nil
}
func (s *stubProvider) ChatCompletionStream(_ context.Context, _ *ChatCompletionRequest) (io.ReadCloser, error) {
	return nil, nil
}

func TestRegistry_ExactMatch(t *testing.T) {
	r := NewRegistry()
	oai := &stubProvider{name: "openai"}
	r.Register(oai, "gpt-4o", "gpt-3.5-turbo")

	p, err := r.Resolve("gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("got provider %q, want %q", p.Name(), "openai")
	}
}

func TestRegistry_PrefixMatch(t *testing.T) {
	r := NewRegistry()
	oai := &stubProvider{name: "openai"}
	r.Register(oai, "gpt-4*")

	tests := []struct {
		model   string
		wantErr bool
	}{
		{"gpt-4o", false},
		{"gpt-4-turbo-preview", false},
		{"gpt-4", false},
		{"claude-3-opus", true},
	}

	for _, tc := range tests {
		t.Run(tc.model, func(t *testing.T) {
			p, err := r.Resolve(tc.model)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for model %q, got provider %q", tc.model, p.Name())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name() != "openai" {
				t.Errorf("got provider %q, want %q", p.Name(), "openai")
			}
		})
	}
}

func TestRegistry_ExactTakesPrecedenceOverPrefix(t *testing.T) {
	r := NewRegistry()
	generic := &stubProvider{name: "generic-gpt4"}
	specific := &stubProvider{name: "specific-gpt4o"}

	r.Register(generic, "gpt-4*")
	r.Register(specific, "gpt-4o")

	p, err := r.Resolve("gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "specific-gpt4o" {
		t.Errorf("exact match should win: got %q, want %q", p.Name(), "specific-gpt4o")
	}

	p2, err := r.Resolve("gpt-4-turbo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p2.Name() != "generic-gpt4" {
		t.Errorf("prefix should match: got %q, want %q", p2.Name(), "generic-gpt4")
	}
}

func TestRegistry_MultipleProviders(t *testing.T) {
	r := NewRegistry()
	oai := &stubProvider{name: "openai"}
	anth := &stubProvider{name: "anthropic"}

	r.Register(oai, "gpt-4*", "gpt-3.5*", "o1*")
	r.Register(anth, "claude*")

	tests := []struct {
		model    string
		wantName string
	}{
		{"gpt-4o", "openai"},
		{"gpt-3.5-turbo", "openai"},
		{"o1-mini", "openai"},
		{"claude-3-5-sonnet-20241022", "anthropic"},
		{"claude-3-haiku-20240307", "anthropic"},
	}

	for _, tc := range tests {
		t.Run(tc.model, func(t *testing.T) {
			p, err := r.Resolve(tc.model)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name() != tc.wantName {
				t.Errorf("got %q, want %q", p.Name(), tc.wantName)
			}
		})
	}
}

func TestRegistry_UnknownModel_ReturnsError(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubProvider{name: "openai"}, "gpt-4*")

	_, err := r.Resolve("llama-3-70b")
	if err == nil {
		t.Error("expected error for unregistered model, got nil")
	}
}
