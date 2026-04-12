# Spec 14: Additional OpenAI-Compatible Providers

## Status
pending

## Goal
Register Groq, Mistral AI, Together AI, Fireworks AI, DeepSeek, and Perplexity as named providers. All use the OpenAI-compatible API format, so after spec 11 (generic OpenAI-compatible provider) is in place, this spec is primarily configuration and well-known defaults.

## Background
Spec 11 implements `OpenAICompatProvider` and `WellKnownBaseURLs`. This spec adds named provider entries and any provider-specific quirks.

## Requirements

### 1. Extend `WellKnownBaseURLs` (if not already done in spec 11)

Ensure these base URLs are in the map:
```go
"groq":         "https://api.groq.com/openai/v1",
"mistral":      "https://api.mistral.ai/v1",
"together":     "https://api.together.xyz/v1",
"fireworks":    "https://api.fireworks.ai/inference/v1",
"deepseek":     "https://api.deepseek.com/v1",
"perplexity":   "https://api.perplexity.ai",
"openrouter":   "https://openrouter.ai/api/v1",
```

### 2. Config section per provider (`pkg/config/config.go`)

```go
Groq struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"groq"`

Mistral struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"mistral"`

Together struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"together"`

Fireworks struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"fireworks"`

DeepSeek struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"deepseek"`

Perplexity struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"perplexity"`
```

Each maps to `LLMGOPHER_GROQ_API_KEY`, etc.

### 3. Provider registration (`cmd/gateway/main.go`)

For each provider with a non-empty API key, create a `NewOpenAICompatProvider(name, baseURL, apiKey)` and register it:

```go
for name, baseURL := range proxy.WellKnownBaseURLs {
    apiKey := getConfigKeyForProvider(cfg, name) // reads the config field
    if apiKey != "" {
        p := proxy.NewOpenAICompatProvider(name, baseURL, apiKey)
        registry.Register(name, p)
    }
}
```

### 4. Provider-specific quirks

Each provider has minor differences; handle them in the generic provider with per-provider hooks if needed:

**Groq:**
- Supports a `reasoning_effort` field in some models — pass through as extra JSON (use `json.RawMessage` in request for passthrough)
- Returns `x-ratelimit-*` headers; log them at debug level

**Mistral:**
- Tool calls use the same format as OpenAI — no translation needed
- Model names: `mistral-large-latest`, `mistral-small-latest`, `codestral-latest`

**Together AI:**
- Model names use `/` separator: `meta-llama/Llama-3-70b-chat-hf`
- Streaming response may include a `usage` block in the final chunk — pass through as-is

**DeepSeek:**
- `deepseek-reasoner` model returns `reasoning_content` in the response — log it but strip from the canonical response to avoid client confusion

**Perplexity:**
- Appends `citations` to responses — strip from canonical response, log at debug level

**OpenRouter:**
- Requires `HTTP-Referer` and `X-Title` headers for attribution — add as optional config fields

For all provider-specific extra fields: accept them in the response but do not include them in the canonical `ChatCompletionResponse` unless they map to existing fields.

### 5. Pricing seed data

Add pricing entries to `internal/storage/migrations/` or the seed data for these providers' popular models so cost tracking works out of the box.

### 6. Default model prefix patterns

Update `internal/proxy/provider_defaults.go` with model prefix patterns that auto-route to each provider without needing explicit model configuration in the DB:

```go
var DefaultModelPrefixes = map[string]string{
    "llama":      "groq",       // or together — let user configure
    "mixtral":    "groq",
    "mistral":    "mistral",
    "deepseek":   "deepseek",
    "sonar":      "perplexity",
}
```

These are fallback prefix hints, not mandatory. Explicit DB model config takes priority.

## Out of Scope
- Custom request format per provider beyond what the generic compat provider handles
- Provider-specific embeddings (Mistral has embeddings — add in a follow-up)
- OpenRouter-specific multi-provider routing (treat as a single provider)

## Acceptance Criteria
- [ ] A Groq API key in config registers a `groq` provider and successfully routes `llama-3.1-70b-versatile`
- [ ] A Mistral API key routes `mistral-large-latest`
- [ ] Together AI routes a Llama model
- [ ] Config uses `LLMGOPHER_GROQ_API_KEY`, etc. env var names
- [ ] Providers registered from DB config (spec 11 pattern) still work
- [ ] DeepSeek `reasoning_content` is stripped from canonical response

## Key Files
- `internal/proxy/provider_defaults.go` — extend well-known base URLs and prefix hints
- `pkg/config/config.go` — add provider config sections
- `cmd/gateway/main.go` — register providers
- `internal/storage/migrations/` — pricing seed data for new providers
