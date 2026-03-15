package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultOpenAIBaseURL    = "https://api.openai.com"
	defaultAnthropicBaseURL = "https://api.anthropic.com"
	defaultGoogleBaseURL    = "https://generativelanguage.googleapis.com"
)

const (
	ErrCodeInvalidAPIKey     = "invalid_api_key"
	ErrCodeQuotaExceeded     = "quota_exceeded"
	ErrCodeNetworkError      = "network_error"
	ErrCodeProviderError     = "provider_error"
	ErrCodeUnsupported       = "unsupported_provider"
	ErrCodeInvalidCredential = "invalid_credential"
)

// ValidationError classifies provider credential validation failures.
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return "credential validation failed"
}

// CredentialValidator validates provider API keys against low-cost endpoints.
type CredentialValidator struct {
	client           *http.Client
	openAIBaseURL    string
	anthropicBaseURL string
	googleGeminiBase string
}

// NewCredentialValidator returns a validator with sensible provider defaults.
func NewCredentialValidator(client *http.Client) *CredentialValidator {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &CredentialValidator{
		client:           client,
		openAIBaseURL:    defaultOpenAIBaseURL,
		anthropicBaseURL: defaultAnthropicBaseURL,
		googleGeminiBase: defaultGoogleBaseURL,
	}
}

// SetOpenAIBaseURLForTest overrides the OpenAI base URL in tests.
func (v *CredentialValidator) SetOpenAIBaseURLForTest(baseURL string) {
	v.openAIBaseURL = strings.TrimSpace(baseURL)
}

// SetAnthropicBaseURLForTest overrides the Anthropic base URL in tests.
func (v *CredentialValidator) SetAnthropicBaseURLForTest(baseURL string) {
	v.anthropicBaseURL = strings.TrimSpace(baseURL)
}

// SetGoogleBaseURLForTest overrides the Google base URL in tests.
func (v *CredentialValidator) SetGoogleBaseURLForTest(baseURL string) {
	v.googleGeminiBase = strings.TrimSpace(baseURL)
}

func (v *CredentialValidator) Validate(ctx context.Context, provider string, apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		return &ValidationError{
			Code:    ErrCodeInvalidCredential,
			Message: "apiKey is required",
		}
	}

	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return v.validateOpenAI(ctx, apiKey)
	case "anthropic":
		return v.validateAnthropic(ctx, apiKey)
	case "google":
		return v.validateGoogle(ctx, apiKey)
	default:
		return &ValidationError{
			Code:    ErrCodeUnsupported,
			Message: "provider must be one of: openai, anthropic, google",
		}
	}
}

func (v *CredentialValidator) validateOpenAI(ctx context.Context, apiKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(v.openAIBaseURL, "/")+"/v1/models", nil)
	if err != nil {
		return &ValidationError{Code: ErrCodeProviderError, Message: "failed to build OpenAI validation request"}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	return v.perform(req)
}

func (v *CredentialValidator) validateAnthropic(ctx context.Context, apiKey string) error {
	body, err := json.Marshal(map[string]any{
		"model":      "claude-3-5-haiku-latest",
		"max_tokens": 1,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
	})
	if err != nil {
		return &ValidationError{Code: ErrCodeProviderError, Message: "failed to build Anthropic validation payload"}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(v.anthropicBaseURL, "/")+"/v1/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return &ValidationError{Code: ErrCodeProviderError, Message: "failed to build Anthropic validation request"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")
	return v.perform(req)
}

func (v *CredentialValidator) validateGoogle(ctx context.Context, apiKey string) error {
	baseURL := strings.TrimRight(v.googleGeminiBase, "/")
	target := baseURL + "/v1beta/models/gemini-pro?key=" + url.QueryEscape(apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return &ValidationError{Code: ErrCodeProviderError, Message: "failed to build Google validation request"}
	}
	return v.perform(req)
}

func (v *CredentialValidator) perform(req *http.Request) error {
	resp, err := v.client.Do(req)
	if err != nil {
		return &ValidationError{Code: ErrCodeNetworkError, Message: "Network Error"}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	normalizedBody := strings.ToLower(string(body))

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &ValidationError{Code: ErrCodeInvalidAPIKey, Message: "Invalid API Key"}
	case http.StatusTooManyRequests:
		return &ValidationError{Code: ErrCodeQuotaExceeded, Message: "Quota Exceeded"}
	default:
		if strings.Contains(normalizedBody, "invalid api key") ||
			strings.Contains(normalizedBody, "incorrect api key") ||
			strings.Contains(normalizedBody, "invalid x-api-key") ||
			strings.Contains(normalizedBody, "api key not valid") {
			return &ValidationError{Code: ErrCodeInvalidAPIKey, Message: "Invalid API Key"}
		}
		if strings.Contains(normalizedBody, "quota") ||
			strings.Contains(normalizedBody, "rate limit") {
			return &ValidationError{Code: ErrCodeQuotaExceeded, Message: "Quota Exceeded"}
		}
		return &ValidationError{
			Code:    ErrCodeProviderError,
			Message: fmt.Sprintf("Validation failed with provider status %d", resp.StatusCode),
		}
	}
}

// AsValidationError extracts ValidationError from wrapped errors.
func AsValidationError(err error) (*ValidationError, bool) {
	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		return nil, false
	}
	return vErr, true
}
