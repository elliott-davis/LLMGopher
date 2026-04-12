package proxy

import "strings"

// WellKnownBaseURLs provides sensible defaults for common OpenAI-compatible
// providers when no explicit base URL is configured.
var WellKnownBaseURLs = map[string]string{
	"groq":       "https://api.groq.com/openai/v1",
	"mistral":    "https://api.mistral.ai/v1",
	"together":   "https://api.together.xyz/v1",
	"fireworks":  "https://api.fireworks.ai/inference/v1",
	"deepseek":   "https://api.deepseek.com/v1",
	"perplexity": "https://api.perplexity.ai",
	"ollama":     "http://localhost:11434/v1",
}

// ResolveProviderBaseURL returns the configured base URL if present; otherwise,
// it falls back to a well-known default keyed by provider name.
func ResolveProviderBaseURL(providerName, configuredBaseURL string) string {
	if trimmed := strings.TrimSpace(configuredBaseURL); trimmed != "" {
		return trimmed
	}

	if defaultURL, ok := WellKnownBaseURLs[strings.ToLower(strings.TrimSpace(providerName))]; ok {
		return defaultURL
	}
	return ""
}
