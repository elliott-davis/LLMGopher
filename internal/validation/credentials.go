package validation

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
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
	return v.ValidateWithBaseURL(ctx, provider, apiKey, "")
}

// ValidateWithBaseURL validates provider credentials and supports provider types
// that require requesting a custom base URL (for example openai_compat).
func (v *CredentialValidator) ValidateWithBaseURL(ctx context.Context, provider string, apiKey string, baseURL string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	apiKey = strings.TrimSpace(apiKey)
	baseURL = strings.TrimSpace(baseURL)

	if provider == "openai_compat" && isLocalhostBaseURL(baseURL) {
		// Local OpenAI-compatible servers often do not require credentials.
		return nil
	}

	if strings.TrimSpace(apiKey) == "" {
		return &ValidationError{
			Code:    ErrCodeInvalidCredential,
			Message: "apiKey is required",
		}
	}

	switch provider {
	case "openai":
		return v.validateOpenAI(ctx, apiKey)
	case "openai_compat":
		return v.validateOpenAICompat(ctx, apiKey, baseURL)
	case "anthropic":
		return v.validateAnthropic(ctx, apiKey)
	case "google":
		return v.validateGoogle(ctx, apiKey)
	default:
		return &ValidationError{
			Code:    ErrCodeUnsupported,
			Message: "provider must be one of: openai, openai_compat, anthropic, google",
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

func (v *CredentialValidator) validateOpenAICompat(ctx context.Context, apiKey string, baseURL string) error {
	targetBaseURL := strings.TrimSpace(baseURL)
	if targetBaseURL == "" {
		targetBaseURL = strings.TrimRight(v.openAIBaseURL, "/") + "/v1"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(targetBaseURL, "/")+"/models", nil)
	if err != nil {
		return &ValidationError{Code: ErrCodeProviderError, Message: "failed to build OpenAI-compatible validation request"}
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

// LoadProviderCredentialTokens returns decrypted provider credential payloads
// for providers configured with bearer-like auth or aws_bedrock auth.
func LoadProviderCredentialTokens(ctx context.Context, db *sql.DB, encryptionKey []byte) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string)
	if db == nil {
		return result, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, credential_ciphertext, credential_nonce
		FROM providers
		WHERE has_credentials = TRUE
		  AND lower(auth_type) IN ('bearer', 'openai_compat', 'aws_bedrock')
	`)
	if err != nil {
		return nil, fmt.Errorf("query provider credentials: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			rawID      string
			ciphertext []byte
			nonce      []byte
		)
		if err := rows.Scan(&rawID, &ciphertext, &nonce); err != nil {
			return nil, fmt.Errorf("scan provider credential: %w", err)
		}

		providerID, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("parse provider id %q: %w", rawID, err)
		}
		if len(ciphertext) == 0 || len(nonce) == 0 {
			continue
		}
		if len(encryptionKey) != 32 {
			return nil, fmt.Errorf("provider credential key must be configured (32 bytes) to decrypt provider %q", rawID)
		}

		token, err := decryptProviderCredential(encryptionKey, ciphertext, nonce)
		if err != nil {
			return nil, fmt.Errorf("decrypt provider credential for %q: %w", rawID, err)
		}
		result[providerID] = strings.TrimSpace(token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provider credentials: %w", err)
	}
	return result, nil
}

func decryptProviderCredential(key []byte, ciphertext []byte, nonce []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(nonce) != gcm.NonceSize() {
		return "", fmt.Errorf("invalid nonce size: got %d want %d", len(nonce), gcm.NonceSize())
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func isLocalhostBaseURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}

	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return false
	}
	if host == "localhost" || host == "::1" {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// AsValidationError extracts ValidationError from wrapped errors.
func AsValidationError(err error) (*ValidationError, bool) {
	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		return nil, false
	}
	return vErr, true
}
