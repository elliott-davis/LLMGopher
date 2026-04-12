package validation

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCredentialValidatorValidateOpenAI(t *testing.T) {
	t.Parallel()

	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/v1/models")
		}
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	validator := NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
	validator.openAIBaseURL = srv.URL

	if err := validator.Validate(context.Background(), "openai", "sk-valid"); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if authHeader != "Bearer sk-valid" {
		t.Fatalf("Authorization header = %q, want %q", authHeader, "Bearer sk-valid")
	}
}

func TestCredentialValidatorValidateAnthropic(t *testing.T) {
	t.Parallel()

	var (
		method           string
		path             string
		apiKeyHeader     string
		versionHeader    string
		contentType      string
		requestBodyModel string
		requestBodyMax   float64
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		apiKeyHeader = r.Header.Get("X-API-Key")
		versionHeader = r.Header.Get("Anthropic-Version")
		contentType = r.Header.Get("Content-Type")

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		requestBodyModel, _ = payload["model"].(string)
		requestBodyMax, _ = payload["max_tokens"].(float64)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"msg_1","content":[{"type":"text","text":"ok"}]}`))
	}))
	defer srv.Close()

	validator := NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
	validator.anthropicBaseURL = srv.URL

	if err := validator.Validate(context.Background(), "anthropic", "anthropic-key"); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}

	if method != http.MethodPost {
		t.Fatalf("method = %q, want %q", method, http.MethodPost)
	}
	if path != "/v1/messages" {
		t.Fatalf("path = %q, want %q", path, "/v1/messages")
	}
	if apiKeyHeader != "anthropic-key" {
		t.Fatalf("X-API-Key = %q, want %q", apiKeyHeader, "anthropic-key")
	}
	if versionHeader != "2023-06-01" {
		t.Fatalf("Anthropic-Version = %q, want %q", versionHeader, "2023-06-01")
	}
	if contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}
	if requestBodyModel != "claude-3-5-haiku-latest" {
		t.Fatalf("model = %q, want %q", requestBodyModel, "claude-3-5-haiku-latest")
	}
	if requestBodyMax != 1 {
		t.Fatalf("max_tokens = %v, want 1", requestBodyMax)
	}
}

func TestCredentialValidatorValidateGoogle(t *testing.T) {
	t.Parallel()

	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"models/gemini-pro"}`))
	}))
	defer srv.Close()

	validator := NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
	validator.googleGeminiBase = srv.URL

	if err := validator.Validate(context.Background(), "google", "google-key"); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if receivedPath != "/v1beta/models/gemini-pro?key=google-key" {
		t.Fatalf("request path = %q, want %q", receivedPath, "/v1beta/models/gemini-pro?key=google-key")
	}
}

func TestCredentialValidatorClassifiesErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		status      int
		body        string
		wantErrCode string
	}{
		{name: "invalid key from status", status: http.StatusUnauthorized, body: `unauthorized`, wantErrCode: ErrCodeInvalidAPIKey},
		{name: "quota from status", status: http.StatusTooManyRequests, body: `quota exceeded`, wantErrCode: ErrCodeQuotaExceeded},
		{name: "invalid key from body", status: http.StatusBadRequest, body: `incorrect api key provided`, wantErrCode: ErrCodeInvalidAPIKey},
		{name: "quota from body", status: http.StatusBadRequest, body: `quota exceeded`, wantErrCode: ErrCodeQuotaExceeded},
		{name: "other provider error", status: http.StatusBadGateway, body: `unexpected`, wantErrCode: ErrCodeProviderError},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			validator := NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
			validator.openAIBaseURL = srv.URL

			err := validator.Validate(context.Background(), "openai", "sk-fail")
			if err == nil {
				t.Fatalf("Validate() error = nil, want non-nil")
			}

			vErr, ok := AsValidationError(err)
			if !ok {
				t.Fatalf("error type = %T, want ValidationError", err)
			}
			if vErr.Code != tc.wantErrCode {
				t.Fatalf("error code = %q, want %q", vErr.Code, tc.wantErrCode)
			}
		})
	}
}

func TestCredentialValidatorNetworkError(t *testing.T) {
	t.Parallel()

	validator := NewCredentialValidator(&http.Client{Timeout: 1 * time.Millisecond})
	validator.openAIBaseURL = "http://127.0.0.1:1"

	err := validator.Validate(context.Background(), "openai", "sk-key")
	if err == nil {
		t.Fatalf("Validate() error = nil, want non-nil")
	}
	vErr, ok := AsValidationError(err)
	if !ok {
		t.Fatalf("error type = %T, want ValidationError", err)
	}
	if vErr.Code != ErrCodeNetworkError {
		t.Fatalf("error code = %q, want %q", vErr.Code, ErrCodeNetworkError)
	}
	if !strings.EqualFold(vErr.Message, "Network Error") {
		t.Fatalf("error message = %q, want %q", vErr.Message, "Network Error")
	}
}

func TestCredentialValidatorValidateOpenAICompat(t *testing.T) {
	t.Parallel()

	var (
		gotPath string
		gotAuth string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	validator := NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
	if err := validator.validateOpenAICompat(context.Background(), "compat-key", srv.URL+"/v1"); err != nil {
		t.Fatalf("validateOpenAICompat() error = %v, want nil", err)
	}
	if gotPath != "/v1/models" {
		t.Fatalf("path = %q, want %q", gotPath, "/v1/models")
	}
	if gotAuth != "Bearer compat-key" {
		t.Fatalf("Authorization header = %q, want %q", gotAuth, "Bearer compat-key")
	}
}

func TestCredentialValidatorValidateOpenAICompatSkipsLocalhost(t *testing.T) {
	t.Parallel()

	validator := NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
	if err := validator.ValidateWithBaseURL(context.Background(), "openai_compat", "", "http://localhost:11434/v1"); err != nil {
		t.Fatalf("ValidateWithBaseURL() error = %v, want nil for localhost", err)
	}
}

func TestLoadProviderCredentialTokens(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	providerID := uuid.New()
	key := mustGenerateCredentialKey(t)
	ciphertext, nonce := mustEncryptCredentialForTest(t, key, []byte("sk-test"))

	rows := sqlmock.NewRows([]string{"id", "credential_ciphertext", "credential_nonce"}).
		AddRow(providerID.String(), ciphertext, nonce)
	mock.ExpectQuery("SELECT id, credential_ciphertext, credential_nonce\\s+FROM providers\\s+WHERE has_credentials = TRUE\\s+AND lower\\(auth_type\\) IN \\('bearer', 'openai_compat', 'aws_bedrock'\\)").
		WillReturnRows(rows)

	got, err := LoadProviderCredentialTokens(context.Background(), db, key)
	if err != nil {
		t.Fatalf("LoadProviderCredentialTokens() error = %v, want nil", err)
	}
	if got[providerID] != "sk-test" {
		t.Fatalf("credential token = %q, want %q", got[providerID], "sk-test")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

func mustGenerateCredentialKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand.Read(key) error = %v", err)
	}
	return key
}

func mustEncryptCredentialForTest(t *testing.T, key []byte, plaintext []byte) (ciphertext []byte, nonce []byte) {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher() error = %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM() error = %v", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("rand.Read(nonce) error = %v", err)
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce
}
