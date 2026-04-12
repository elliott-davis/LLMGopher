package proxy

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestRegisterDynamicOpenAICompatProviders_RegistersBearerAndNone(t *testing.T) {
	t.Parallel()

	registry := llm.NewRegistry()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	groqID := uuid.New()
	ollamaID := uuid.New()
	state := &storage.GatewayState{
		Providers: map[uuid.UUID]*llm.ProviderConfig{
			groqID: {
				ID:       groqID.String(),
				Name:     "Groq",
				BaseURL:  "",
				AuthType: "bearer",
			},
			ollamaID: {
				ID:       ollamaID.String(),
				Name:     "Ollama",
				BaseURL:  "",
				AuthType: "none",
			},
		},
	}

	RegisterDynamicOpenAICompatProviders(registry, state, map[uuid.UUID]string{
		groqID: "sk-groq",
	}, logger)

	groqProvider, err := registry.ResolveProvider("groq")
	if err != nil {
		t.Fatalf("ResolveProvider(groq) error = %v, want nil", err)
	}
	if groqProvider.Name() != "Groq" {
		t.Fatalf("provider name = %q, want %q", groqProvider.Name(), "Groq")
	}

	ollamaProvider, err := registry.ResolveProvider("ollama")
	if err != nil {
		t.Fatalf("ResolveProvider(ollama) error = %v, want nil", err)
	}
	if ollamaProvider.Name() != "Ollama" {
		t.Fatalf("provider name = %q, want %q", ollamaProvider.Name(), "Ollama")
	}
}

func TestRegisterDynamicOpenAICompatProviders_RefreshAddsProvider(t *testing.T) {
	t.Parallel()

	registry := llm.NewRegistry()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	groqID := uuid.New()
	state := &storage.GatewayState{
		Providers: map[uuid.UUID]*llm.ProviderConfig{
			groqID: {
				ID:       groqID.String(),
				Name:     "Groq",
				BaseURL:  "",
				AuthType: "bearer",
			},
		},
	}

	RegisterDynamicOpenAICompatProviders(registry, state, map[uuid.UUID]string{
		groqID: "sk-groq",
	}, logger)

	if _, err := registry.ResolveProvider("together"); err == nil {
		t.Fatal("ResolveProvider(together) error = nil, want non-nil before refresh")
	}

	togetherID := uuid.New()
	state.Providers[togetherID] = &llm.ProviderConfig{
		ID:       togetherID.String(),
		Name:     "Together",
		BaseURL:  "",
		AuthType: "openai_compat",
	}
	RegisterDynamicOpenAICompatProviders(registry, state, map[uuid.UUID]string{
		groqID:     "sk-groq",
		togetherID: "sk-together",
	}, logger)

	if _, err := registry.ResolveProvider("together"); err != nil {
		t.Fatalf("ResolveProvider(together) error = %v, want nil after refresh", err)
	}
}

func TestRegisterDynamicOpenAICompatProviders_RegistersBedrockProvider(t *testing.T) {
	t.Parallel()

	registry := llm.NewRegistry()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	bedrockID := uuid.New()
	state := &storage.GatewayState{
		Providers: map[uuid.UUID]*llm.ProviderConfig{
			bedrockID: {
				ID:       bedrockID.String(),
				Name:     "AWS Bedrock",
				BaseURL:  "us-east-1",
				AuthType: "aws_bedrock",
			},
		},
	}

	RegisterDynamicOpenAICompatProviders(registry, state, map[uuid.UUID]string{
		bedrockID: `{"access_key_id":"AKIAX","secret_access_key":"secret"}`,
	}, logger)

	provider, err := registry.ResolveProvider("bedrock")
	if err != nil {
		t.Fatalf("ResolveProvider(bedrock) error = %v, want nil", err)
	}
	if provider.Name() != "bedrock" {
		t.Fatalf("provider name = %q, want %q", provider.Name(), "bedrock")
	}
}
