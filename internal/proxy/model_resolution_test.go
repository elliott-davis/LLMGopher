package proxy

import (
	"testing"

	"github.com/google/uuid"

	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/pkg/llm"
)

func TestResolveConfiguredModel_ByAlias(t *testing.T) {
	state := testGatewayState("vertex", "vertex_ai/gemini-2.5-flash", "gemini-2.5-flash")

	got, ok := resolveConfiguredModel(state, "vertex_ai/gemini-2.5-flash")
	if !ok {
		t.Fatal("resolveConfiguredModel() matched = false, want true")
	}
	if got.Alias != "vertex_ai/gemini-2.5-flash" {
		t.Fatalf("alias = %q, want %q", got.Alias, "vertex_ai/gemini-2.5-flash")
	}
}

func TestResolveConfiguredModel_ByProviderQualifiedName(t *testing.T) {
	state := testGatewayState("vertex", "vertex_ai/gemini-2.5-flash", "gemini-2.5-flash")

	got, ok := resolveConfiguredModel(state, "vertex/gemini-2.5-flash")
	if !ok {
		t.Fatal("resolveConfiguredModel() matched = false, want true")
	}
	if got.Name != "gemini-2.5-flash" {
		t.Fatalf("name = %q, want %q", got.Name, "gemini-2.5-flash")
	}
}

func TestResolveConfiguredModel_ByProviderQualifiedNameWithPublisherModel(t *testing.T) {
	state := testGatewayState("vertex", "vertex_ai/google-gemini", "google/gemini-2.5-flash")

	got, ok := resolveConfiguredModel(state, "vertex/gemini-2.5-flash")
	if !ok {
		t.Fatal("resolveConfiguredModel() matched = false, want true")
	}
	if got.Name != "google/gemini-2.5-flash" {
		t.Fatalf("name = %q, want %q", got.Name, "google/gemini-2.5-flash")
	}
}

func TestResolveConfiguredModel_ByProviderQualifiedNameWithDisplayProviderName(t *testing.T) {
	state := testGatewayStateWithProvider(&llm.ProviderConfig{
		ID:       uuid.NewString(),
		Name:     "Google Vertex",
		BaseURL:  "https://us-central1-aiplatform.googleapis.com/v1beta1",
		AuthType: "vertex_service_account",
	}, "gemini-2.5-flash", "gemini-2.5-flash")

	got, ok := resolveConfiguredModel(state, "vertex/gemini-2.5-flash")
	if !ok {
		t.Fatal("resolveConfiguredModel() matched = false, want true")
	}
	if got.Name != "gemini-2.5-flash" {
		t.Fatalf("name = %q, want %q", got.Name, "gemini-2.5-flash")
	}
}

func TestResolveConfiguredModel_UnknownProviderQualifiedName(t *testing.T) {
	state := testGatewayState("vertex", "vertex_ai/gemini-2.5-flash", "gemini-2.5-flash")

	if _, ok := resolveConfiguredModel(state, "openai/gemini-2.5-flash"); ok {
		t.Fatal("resolveConfiguredModel() matched = true, want false")
	}
}

func TestPreferredProviderRegistryName_InfersVertexFromProviderConfig(t *testing.T) {
	got := preferredProviderRegistryName(&llm.ProviderConfig{
		Name:     "Google Vertex",
		BaseURL:  "https://us-central1-aiplatform.googleapis.com/v1beta1",
		AuthType: "vertex_service_account",
	}, "gemini-2.5-flash")

	if got != "vertex" {
		t.Fatalf("preferredProviderRegistryName() = %q, want %q", got, "vertex")
	}
}

func TestPreferredProviderRegistryName_UsesRequestedPrefix(t *testing.T) {
	got := preferredProviderRegistryName(&llm.ProviderConfig{
		Name:     "Google Vertex",
		BaseURL:  "https://us-central1-aiplatform.googleapis.com/v1beta1",
		AuthType: "vertex_service_account",
	}, "vertex/gemini-2.5-flash")

	if got != "vertex" {
		t.Fatalf("preferredProviderRegistryName() = %q, want %q", got, "vertex")
	}
}

func TestPreferredProviderRegistryName_PassthroughCustomProviderName(t *testing.T) {
	got := preferredProviderRegistryName(&llm.ProviderConfig{
		Name:     "Groq",
		BaseURL:  "https://api.groq.com/openai/v1",
		AuthType: "openai_compat",
	}, "llama3")

	if got != "Groq" {
		t.Fatalf("preferredProviderRegistryName() = %q, want %q", got, "Groq")
	}
}

func testGatewayState(providerName, alias, name string) *storage.GatewayState {
	return testGatewayStateWithProvider(&llm.ProviderConfig{
		ID:   uuid.NewString(),
		Name: providerName,
	}, alias, name)
}

func testGatewayStateWithProvider(provider *llm.ProviderConfig, alias, name string) *storage.GatewayState {
	providerID := uuid.New()
	provider.ID = providerID.String()
	return &storage.GatewayState{
		APIKeys: map[string]*llm.APIKey{},
		Models: map[string]*llm.Model{
			alias: {
				ID:         uuid.NewString(),
				ProviderID: providerID.String(),
				Name:       name,
				Alias:      alias,
			},
		},
		Providers: map[uuid.UUID]*llm.ProviderConfig{
			providerID: provider,
		},
	}
}
