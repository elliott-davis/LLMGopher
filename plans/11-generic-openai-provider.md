# Spec 11: Generic OpenAI-Compatible Provider

## Status
pending

## Goal
Implement a provider that forwards requests to any OpenAI-compatible endpoint by configuring only a `base_url` and API key. This single implementation covers Groq, Mistral AI, Ollama, vLLM, LM Studio, Together AI, Fireworks AI, DeepSeek, Perplexity, and any other server implementing the OpenAI chat completions API.

## Background
`internal/proxy/provider_openai.go` implements the OpenAI provider by forwarding the request to `api.openai.com`. The implementation is essentially a configurable HTTP client with the OpenAI base URL hardcoded. The generic provider is a parameterized version of this.

Providers are registered in `cmd/gateway/main.go` and stored in the `ProviderRegistry`. The `llm.Provider` interface requires `Name()`, `ChatCompletion()`, and `ChatCompletionStream()`.

The `providers` table stores `base_url` and `auth_type`. Credentials are retrieved via `internal/validation/credentials.go`.

## Requirements

### 1. New provider (`internal/proxy/provider_openai_compat.go`)

```go
type OpenAICompatProvider struct {
    name    string
    baseURL string
    apiKey  string
    client  *http.Client
}

func NewOpenAICompatProvider(name, baseURL, apiKey string) *OpenAICompatProvider
func (p *OpenAICompatProvider) Name() string
func (p *OpenAICompatProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error)
func (p *OpenAICompatProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error)
```

Implementation is identical to `provider_openai.go` except `baseURL` is a constructor parameter instead of a hardcoded constant. The request is serialized to JSON and POSTed to `{baseURL}/chat/completions`.

For auth: if `apiKey` is non-empty, set `Authorization: Bearer {apiKey}`. If empty, omit the header (some local servers like Ollama don't require auth).

### 2. Provider factory in `cmd/gateway/main.go`

When loading providers from the state cache at startup (and on cache refresh), for each provider row with `auth_type = "bearer"` or `auth_type = "openai_compat"`:
1. Retrieve the decrypted API key from the credentials store
2. Construct a `NewOpenAICompatProvider(providerConfig.Name, providerConfig.BaseURL, apiKey)`
3. Register it in the registry using `providerConfig.Name` as the key

This means DB-configured providers that aren't OpenAI/Anthropic/Vertex are automatically handled by this generic provider.

### 3. Provider name resolution

In `internal/proxy/model_resolution.go`, `preferredProviderRegistryName` maps provider configs to registry keys. Add a case: if the provider config name is not a recognized built-in, return the config name as-is (so it matches the generic provider registered with that name).

### 4. Well-known base URLs as defaults

Add a helper `internal/proxy/provider_defaults.go` with a map of well-known provider names to base URLs:
```go
var WellKnownBaseURLs = map[string]string{
    "groq":        "https://api.groq.com/openai/v1",
    "mistral":     "https://api.mistral.ai/v1",
    "together":    "https://api.together.xyz/v1",
    "fireworks":   "https://api.fireworks.ai/inference/v1",
    "deepseek":    "https://api.deepseek.com/v1",
    "perplexity":  "https://api.perplexity.ai",
    "ollama":      "http://localhost:11434/v1",
}
```

If a provider's `base_url` is empty and its name matches a well-known key, use the default. This allows operators to add a provider by name alone.

### 5. Credential validation (`internal/validation/credentials.go`)

Add an `"openai_compat"` case to the validator that tests the configured base URL with a minimal request (e.g., `GET {baseURL}/models` or a minimal chat completion). Fall back to skipping validation if the base URL is a localhost address.

## Out of Scope
- Embeddings support via the generic provider (add if needed; OpenAI compat servers often support `/embeddings` too, but keep this spec focused)
- Streaming format differences (some compat servers have minor SSE deviations — handle in a follow-up)
- Tool/function call translation (generic compat servers use the same format as OpenAI)

## Acceptance Criteria
- [ ] A provider with `base_url: "https://api.groq.com/openai/v1"` and a Groq API key successfully completes a chat request
- [ ] An Ollama provider with no API key (`auth_type: "none"`) works with a locally running Ollama server
- [ ] Provider registered via DB (not static config) is picked up at startup and after state cache refresh
- [ ] Well-known base URL defaulting works for `groq`, `mistral`, `together`
- [ ] Model routing routes to the correct generic provider by provider name
- [ ] Unit test covers the provider with a mock HTTP server

## Key Files
- `internal/proxy/provider_openai_compat.go` — new file
- `internal/proxy/provider_defaults.go` — well-known base URLs
- `cmd/gateway/main.go` — dynamic provider registration from state cache
- `internal/proxy/model_resolution.go` — generic provider name passthrough
- `internal/validation/credentials.go` — openai_compat validation case
