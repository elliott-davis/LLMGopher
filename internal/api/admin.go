package api

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ed007183/llmgopher/internal/storage"
	"github.com/ed007183/llmgopher/internal/validation"
	"github.com/ed007183/llmgopher/pkg/llm"
	"github.com/google/uuid"
)

// HandleGetKeys returns the active API key cache snapshot as a JSON array.
func HandleGetKeys(cache *storage.StateCache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		keys := make([]*llm.APIKey, 0)
		if cache != nil {
			if state := cache.Load(); state != nil {
				keys = make([]*llm.APIKey, 0, len(state.APIKeys))
				for _, key := range state.APIKeys {
					keys = append(keys, key)
				}
			}
		}

		WriteJSON(w, http.StatusOK, keys)
	}
}

// HandleGetModels returns the configured model cache snapshot as a JSON array.
func HandleGetModels(cache *storage.StateCache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		models := make([]*llm.Model, 0)
		if cache != nil {
			if state := cache.Load(); state != nil {
				models = make([]*llm.Model, 0, len(state.Models))
				for _, model := range state.Models {
					models = append(models, model)
				}
			}
		}

		WriteJSON(w, http.StatusOK, models)
	}
}

// HandleGetProviders returns the configured provider cache snapshot as a JSON array.
func HandleGetProviders(cache *storage.StateCache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		providers := make([]*llm.ProviderConfig, 0)
		if cache != nil {
			if state := cache.Load(); state != nil {
				providers = make([]*llm.ProviderConfig, 0, len(state.Providers))
				for _, provider := range state.Providers {
					providers = append(providers, provider)
				}
			}
		}

		WriteJSON(w, http.StatusOK, providers)
	}
}

type auditEntryResponse struct {
	ID           int64     `json:"id"`
	RequestID    string    `json:"request_id"`
	APIKeyID     string    `json:"api_key_id"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	PromptTokens int       `json:"prompt_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	CostUSD      float64   `json:"cost_usd"`
	StatusCode   int       `json:"status_code"`
	LatencyMS    int64     `json:"latency_ms"`
	Streaming    bool      `json:"streaming"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

type getAuditLogResponse struct {
	Data   []auditEntryResponse `json:"data"`
	Total  int                  `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

const (
	auditStatusSuccess = "success"
	auditStatusError   = "error"
	auditDefaultLimit  = 100
	auditMaxLimit      = 1000
)

// HandleGetAuditLog returns paginated audit entries from PostgreSQL.
func HandleGetAuditLog(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		query := storage.AuditQuery{
			APIKeyID: strings.TrimSpace(r.URL.Query().Get("api_key_id")),
			Model:    strings.TrimSpace(r.URL.Query().Get("model")),
			Provider: strings.TrimSpace(r.URL.Query().Get("provider")),
			Limit:    auditDefaultLimit,
			Offset:   0,
		}

		statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
		if statusFilter != "" && statusFilter != auditStatusSuccess && statusFilter != auditStatusError {
			WriteError(w, http.StatusBadRequest, "status must be one of: success, error", "invalid_request_error")
			return
		}
		query.Status = statusFilter

		fromRaw := strings.TrimSpace(r.URL.Query().Get("from"))
		if fromRaw != "" {
			from, err := time.Parse(time.RFC3339, fromRaw)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "from must be an ISO 8601 timestamp", "invalid_request_error")
				return
			}
			query.From = &from
		}

		toRaw := strings.TrimSpace(r.URL.Query().Get("to"))
		if toRaw != "" {
			to, err := time.Parse(time.RFC3339, toRaw)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "to must be an ISO 8601 timestamp", "invalid_request_error")
				return
			}
			query.To = &to
		}

		if query.From != nil && query.To != nil && query.From.After(*query.To) {
			WriteError(w, http.StatusBadRequest, "from must be before or equal to to", "invalid_request_error")
			return
		}

		limitRaw := strings.TrimSpace(r.URL.Query().Get("limit"))
		if limitRaw != "" {
			limit, err := strconv.Atoi(limitRaw)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "limit must be a positive integer", "invalid_request_error")
				return
			}
			if limit <= 0 {
				WriteError(w, http.StatusBadRequest, "limit must be a positive integer", "invalid_request_error")
				return
			}
			if limit > auditMaxLimit {
				limit = auditMaxLimit
			}
			query.Limit = limit
		}

		offsetRaw := strings.TrimSpace(r.URL.Query().Get("offset"))
		if offsetRaw != "" {
			offset, err := strconv.Atoi(offsetRaw)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "offset must be a non-negative integer", "invalid_request_error")
				return
			}
			if offset < 0 {
				WriteError(w, http.StatusBadRequest, "offset must be a non-negative integer", "invalid_request_error")
				return
			}
			query.Offset = offset
		}

		result, err := storage.QueryAuditLog(r.Context(), db, query)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to query audit log"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		data := make([]auditEntryResponse, 0, len(result.Data))
		for _, entry := range result.Data {
			data = append(data, auditEntryResponse{
				ID:           entry.ID,
				RequestID:    entry.RequestID,
				APIKeyID:     entry.APIKeyID,
				Model:        entry.Model,
				Provider:     entry.Provider,
				PromptTokens: entry.PromptTokens,
				OutputTokens: entry.OutputTokens,
				TotalTokens:  entry.TotalTokens,
				CostUSD:      entry.CostUSD,
				StatusCode:   entry.StatusCode,
				LatencyMS:    entry.Latency.Milliseconds(),
				Streaming:    entry.Streaming,
				ErrorMessage: entry.ErrorMessage,
				CreatedAt:    entry.CreatedAt,
			})
		}

		WriteJSON(w, http.StatusOK, getAuditLogResponse{
			Data:   data,
			Total:  result.Total,
			Limit:  query.Limit,
			Offset: query.Offset,
		})
	}
}

type createModelRequest struct {
	Alias         string `json:"alias"`
	Name          string `json:"name"`
	ProviderID    string `json:"provider_id"`
	ContextWindow int    `json:"context_window"`
	RateLimitRPS  int    `json:"rate_limit_rps"`
}

type createProviderRequest struct {
	Name            string `json:"name"`
	BaseURL         string `json:"base_url"`
	AuthType        string `json:"auth_type"`
	CredentialToken string `json:"credential_token,omitempty"`
}

type createAPIKeyRequest struct {
	Name          string            `json:"name"`
	RateLimitRPS  int               `json:"rate_limit_rps"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	AllowedModels []string          `json:"allowed_models,omitempty"`
}

type createAPIKeyResponse struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	RateLimitRPS  int               `json:"rate_limit_rps"`
	IsActive      bool              `json:"is_active"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	AllowedModels []string          `json:"allowed_models,omitempty"`
	APIKey        string            `json:"api_key"`
}

type optionalTime struct {
	Set   bool
	Value *time.Time
}

func (o *optionalTime) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	var parsed time.Time
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	o.Value = &parsed
	return nil
}

type updateAPIKeyRequest struct {
	Name          *string           `json:"name,omitempty"`
	RateLimitRPS  *int              `json:"rate_limit_rps,omitempty"`
	ExpiresAt     optionalTime      `json:"expires_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	AllowedModels []string          `json:"allowed_models,omitempty"`
	IsActive      *bool             `json:"is_active,omitempty"`
}

type upsertBudgetRequest struct {
	BudgetUSD         float64    `json:"budget_usd"`
	AlertThresholdPct *int       `json:"alert_threshold_pct"`
	BudgetDuration    *string    `json:"budget_duration"`
	BudgetResetAt     *time.Time `json:"budget_reset_at"`
}

type budgetResponse struct {
	APIKeyID          string     `json:"api_key_id"`
	BudgetUSD         float64    `json:"budget_usd"`
	SpentUSD          float64    `json:"spent_usd"`
	RemainingUSD      float64    `json:"remaining_usd"`
	AlertThresholdPct *int       `json:"alert_threshold_pct"`
	BudgetDuration    *string    `json:"budget_duration"`
	BudgetResetAt     *time.Time `json:"budget_reset_at"`
}

type validateCredentialRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
	APIKeyV1 string `json:"api_key"`
}

type validateCredentialResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
	Code  string `json:"code,omitempty"`
}

const maxProviderCredentialBytes = 2 << 20 // 2 MiB

func inferProviderForCredentialValidation(baseURL string) string {
	normalized := strings.ToLower(strings.TrimSpace(baseURL))
	switch {
	case strings.Contains(normalized, "api.openai.com"):
		return "openai"
	case strings.Contains(normalized, "api.anthropic.com"):
		return "anthropic"
	case strings.Contains(normalized, "generativelanguage.googleapis.com"),
		strings.Contains(normalized, "googleapis.com"):
		return "google"
	default:
		return ""
	}
}

func writeCredentialValidationFailure(w http.ResponseWriter, err error) {
	if vErr, ok := validation.AsValidationError(err); ok {
		status := http.StatusBadGateway
		switch vErr.Code {
		case validation.ErrCodeInvalidCredential, validation.ErrCodeUnsupported:
			status = http.StatusBadRequest
		case validation.ErrCodeInvalidAPIKey:
			status = http.StatusUnauthorized
		case validation.ErrCodeQuotaExceeded:
			status = http.StatusTooManyRequests
		case validation.ErrCodeNetworkError:
			status = http.StatusServiceUnavailable
		}
		WriteJSON(w, status, validateCredentialResponse{
			Valid: false,
			Error: vErr.Message,
			Code:  vErr.Code,
		})
		return
	}

	WriteJSON(w, http.StatusBadGateway, validateCredentialResponse{
		Valid: false,
		Error: "Validation failed",
		Code:  validation.ErrCodeProviderError,
	})
}

// HandleValidateCredential verifies provider API credentials before persistence.
func HandleValidateCredential(validator *validation.CredentialValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if validator == nil {
			WriteError(w, http.StatusServiceUnavailable, "credential validation is unavailable", "service_unavailable")
			return
		}

		var payload validateCredentialRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
			return
		}

		provider := strings.ToLower(strings.TrimSpace(payload.Provider))
		apiKey := strings.TrimSpace(payload.APIKey)
		if apiKey == "" {
			apiKey = strings.TrimSpace(payload.APIKeyV1)
		}

		if provider == "" || apiKey == "" {
			WriteError(w, http.StatusBadRequest, "provider and apiKey are required", "invalid_request_error")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
		defer cancel()

		if err := validator.Validate(ctx, provider, apiKey); err != nil {
			writeCredentialValidationFailure(w, err)
			return
		}

		WriteJSON(w, http.StatusOK, validateCredentialResponse{Valid: true})
	}
}

// HandleCreateModel inserts a new model row into PostgreSQL.
func HandleCreateModel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		var payload createModelRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
			return
		}

		if strings.TrimSpace(payload.Alias) == "" ||
			strings.TrimSpace(payload.Name) == "" ||
			strings.TrimSpace(payload.ProviderID) == "" ||
			payload.ContextWindow <= 0 ||
			payload.RateLimitRPS < 0 {
			WriteError(w, http.StatusBadRequest, "alias, name, provider_id, positive context_window, and non-negative rate_limit_rps are required", "invalid_request_error")
			return
		}

		if _, err := uuid.Parse(strings.TrimSpace(payload.ProviderID)); err != nil {
			WriteError(w, http.StatusBadRequest, "provider_id must be a valid UUID", "invalid_request_error")
			return
		}

		now := time.Now().UTC()
		_, err := db.ExecContext(
			r.Context(),
			`INSERT INTO models (id, alias, name, provider_id, context_window, rate_limit_rps, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			uuid.NewString(),
			payload.Alias,
			payload.Name,
			payload.ProviderID,
			payload.ContextWindow,
			payload.RateLimitRPS,
			now,
			now,
		)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to create model"
			if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
				status = http.StatusConflict
				msg = "model alias already exists"
			}
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// HandleUpdateModel updates an existing model row by ID.
func HandleUpdateModel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "model id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "model id must be a valid UUID", "invalid_request_error")
			return
		}

		var payload createModelRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
			return
		}

		if strings.TrimSpace(payload.Alias) == "" ||
			strings.TrimSpace(payload.Name) == "" ||
			strings.TrimSpace(payload.ProviderID) == "" ||
			payload.ContextWindow <= 0 ||
			payload.RateLimitRPS < 0 {
			WriteError(w, http.StatusBadRequest, "alias, name, provider_id, positive context_window, and non-negative rate_limit_rps are required", "invalid_request_error")
			return
		}

		if _, err := uuid.Parse(strings.TrimSpace(payload.ProviderID)); err != nil {
			WriteError(w, http.StatusBadRequest, "provider_id must be a valid UUID", "invalid_request_error")
			return
		}

		result, err := db.ExecContext(
			r.Context(),
			`UPDATE models SET alias = $1, name = $2, provider_id = $3, context_window = $4, rate_limit_rps = $5, updated_at = $6 WHERE id = $7`,
			payload.Alias,
			payload.Name,
			payload.ProviderID,
			payload.ContextWindow,
			payload.RateLimitRPS,
			time.Now().UTC(),
			id,
		)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to update model"
			if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
				status = http.StatusConflict
				msg = "model alias already exists"
			}
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			WriteError(w, http.StatusNotFound, "model not found", "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// HandleDeleteModel deletes a model row by ID.
func HandleDeleteModel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "model id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "model id must be a valid UUID", "invalid_request_error")
			return
		}

		result, err := db.ExecContext(r.Context(), `DELETE FROM models WHERE id = $1`, id)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to delete model"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			WriteError(w, http.StatusNotFound, "model not found", "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleCreateProvider inserts a new provider row into PostgreSQL.
func HandleCreateProvider(db *sql.DB, providerCredentialKey []byte) http.HandlerFunc {
	return HandleCreateProviderWithValidation(db, providerCredentialKey, nil)
}

// HandleCreateProviderWithValidation inserts a new provider row with optional credential validation.
func HandleCreateProviderWithValidation(db *sql.DB, providerCredentialKey []byte, validator *validation.CredentialValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		var (
			payload            createProviderRequest
			credentialFileName string
			credentialBytes    []byte
		)

		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(maxProviderCredentialBytes); err != nil {
				WriteError(w, http.StatusBadRequest, "invalid multipart form", "invalid_request_error")
				return
			}

			payload = createProviderRequest{
				Name:            r.FormValue("name"),
				BaseURL:         r.FormValue("base_url"),
				AuthType:        r.FormValue("auth_type"),
				CredentialToken: r.FormValue("credential_token"),
			}

			file, header, err := r.FormFile("credential_file")
			if err == nil {
				defer file.Close()
				credentialFileName = header.Filename
				credentialBytes, err = io.ReadAll(io.LimitReader(file, maxProviderCredentialBytes+1))
				if err != nil {
					WriteError(w, http.StatusBadRequest, "failed to read credential file", "invalid_request_error")
					return
				}
				if len(credentialBytes) > maxProviderCredentialBytes {
					WriteError(w, http.StatusBadRequest, "credential file is too large", "invalid_request_error")
					return
				}
			} else if !errors.Is(err, http.ErrMissingFile) {
				WriteError(w, http.StatusBadRequest, "invalid credential file upload", "invalid_request_error")
				return
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
				return
			}
		}

		if strings.TrimSpace(payload.Name) == "" ||
			strings.TrimSpace(payload.BaseURL) == "" ||
			strings.TrimSpace(payload.AuthType) == "" {
			WriteError(w, http.StatusBadRequest, "name, base_url, and auth_type are required", "invalid_request_error")
			return
		}

		requiresCredentialFile := strings.EqualFold(strings.TrimSpace(payload.AuthType), "vertex_service_account")
		if requiresCredentialFile && len(credentialBytes) == 0 {
			WriteError(w, http.StatusBadRequest, "credential_file is required for auth_type vertex_service_account", "invalid_request_error")
			return
		}
		requiresBearerToken := strings.EqualFold(strings.TrimSpace(payload.AuthType), "bearer")
		if requiresBearerToken && strings.TrimSpace(payload.CredentialToken) == "" {
			WriteError(w, http.StatusBadRequest, "credential_token is required for auth_type bearer", "invalid_request_error")
			return
		}

		credentialToken := strings.TrimSpace(payload.CredentialToken)
		if validator != nil && requiresBearerToken && credentialToken != "" {
			if providerName := inferProviderForCredentialValidation(payload.BaseURL); providerName != "" {
				if err := validator.Validate(r.Context(), providerName, credentialToken); err != nil {
					writeCredentialValidationFailure(w, err)
					return
				}
			}
		}
		if len(credentialBytes) == 0 && credentialToken != "" {
			credentialBytes = []byte(credentialToken)
		}

		var credentialCiphertext, credentialNonce []byte
		hasCredentials := len(credentialBytes) > 0
		if hasCredentials {
			if len(providerCredentialKey) != 32 {
				WriteError(w, http.StatusServiceUnavailable, "credential encryption is not configured", "service_unavailable")
				return
			}
			var err error
			credentialCiphertext, credentialNonce, err = encryptProviderCredential(providerCredentialKey, credentialBytes)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "failed to encrypt credential file", "server_error")
				return
			}
		}

		now := time.Now().UTC()
		_, err := db.ExecContext(
			r.Context(),
			`INSERT INTO providers (
				id, name, base_url, auth_type, created_at, updated_at,
				credential_file_name, credential_ciphertext, credential_nonce, has_credentials
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.NewString(),
			payload.Name,
			payload.BaseURL,
			payload.AuthType,
			now,
			now,
			credentialFileName,
			credentialCiphertext,
			credentialNonce,
			hasCredentials,
		)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to create provider"
			if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
				status = http.StatusConflict
				msg = "provider name already exists"
			}
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// HandleUpdateProvider updates an existing provider row by ID.
func HandleUpdateProvider(db *sql.DB, providerCredentialKey []byte) http.HandlerFunc {
	return HandleUpdateProviderWithValidation(db, providerCredentialKey, nil)
}

// HandleUpdateProviderWithValidation updates an existing provider row with optional credential validation.
func HandleUpdateProviderWithValidation(db *sql.DB, providerCredentialKey []byte, validator *validation.CredentialValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "provider id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "provider id must be a valid UUID", "invalid_request_error")
			return
		}

		var (
			payload            createProviderRequest
			credentialFileName string
			credentialBytes    []byte
		)

		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(maxProviderCredentialBytes); err != nil {
				WriteError(w, http.StatusBadRequest, "invalid multipart form", "invalid_request_error")
				return
			}

			payload = createProviderRequest{
				Name:            r.FormValue("name"),
				BaseURL:         r.FormValue("base_url"),
				AuthType:        r.FormValue("auth_type"),
				CredentialToken: r.FormValue("credential_token"),
			}

			file, header, err := r.FormFile("credential_file")
			if err == nil {
				defer file.Close()
				credentialFileName = header.Filename
				credentialBytes, err = io.ReadAll(io.LimitReader(file, maxProviderCredentialBytes+1))
				if err != nil {
					WriteError(w, http.StatusBadRequest, "failed to read credential file", "invalid_request_error")
					return
				}
				if len(credentialBytes) > maxProviderCredentialBytes {
					WriteError(w, http.StatusBadRequest, "credential file is too large", "invalid_request_error")
					return
				}
			} else if !errors.Is(err, http.ErrMissingFile) {
				WriteError(w, http.StatusBadRequest, "invalid credential file upload", "invalid_request_error")
				return
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
				return
			}
		}

		if strings.TrimSpace(payload.Name) == "" ||
			strings.TrimSpace(payload.BaseURL) == "" ||
			strings.TrimSpace(payload.AuthType) == "" {
			WriteError(w, http.StatusBadRequest, "name, base_url, and auth_type are required", "invalid_request_error")
			return
		}

		credentialToken := strings.TrimSpace(payload.CredentialToken)
		if validator != nil &&
			strings.EqualFold(strings.TrimSpace(payload.AuthType), "bearer") &&
			credentialToken != "" {
			if providerName := inferProviderForCredentialValidation(payload.BaseURL); providerName != "" {
				if err := validator.Validate(r.Context(), providerName, credentialToken); err != nil {
					writeCredentialValidationFailure(w, err)
					return
				}
			}
		}
		if len(credentialBytes) == 0 && credentialToken != "" {
			credentialBytes = []byte(credentialToken)
		}

		credentialProvided := len(credentialBytes) > 0
		var credentialCiphertext, credentialNonce []byte
		if credentialProvided {
			if len(providerCredentialKey) != 32 {
				WriteError(w, http.StatusServiceUnavailable, "credential encryption is not configured", "service_unavailable")
				return
			}
			var err error
			credentialCiphertext, credentialNonce, err = encryptProviderCredential(providerCredentialKey, credentialBytes)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "failed to encrypt credential file", "server_error")
				return
			}
		}

		result, err := db.ExecContext(
			r.Context(),
			`UPDATE providers
			 SET name = $1,
			     base_url = $2,
			     auth_type = $3,
			     updated_at = $4,
			     credential_file_name = CASE WHEN $5 THEN $6 ELSE credential_file_name END,
			     credential_ciphertext = CASE WHEN $5 THEN $7 ELSE credential_ciphertext END,
			     credential_nonce = CASE WHEN $5 THEN $8 ELSE credential_nonce END,
			     has_credentials = CASE WHEN $5 THEN $9 ELSE has_credentials END
			 WHERE id = $10`,
			payload.Name,
			payload.BaseURL,
			payload.AuthType,
			time.Now().UTC(),
			credentialProvided,
			credentialFileName,
			credentialCiphertext,
			credentialNonce,
			credentialProvided,
			id,
		)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to update provider"
			if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
				status = http.StatusConflict
				msg = "provider name already exists"
			}
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			WriteError(w, http.StatusNotFound, "provider not found", "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// HandleDeleteProvider deletes a provider row by ID if no models are attached.
func HandleDeleteProvider(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "provider id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "provider id must be a valid UUID", "invalid_request_error")
			return
		}

		var attachedModels int
		if err := db.QueryRowContext(
			r.Context(),
			`SELECT count(*) FROM models WHERE provider_id = $1`,
			id,
		).Scan(&attachedModels); err != nil {
			status := http.StatusInternalServerError
			msg := "failed to check attached models"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		if attachedModels > 0 {
			WriteJSON(w, http.StatusConflict, map[string]string{
				"error": fmt.Sprintf("Cannot delete provider. It is currently in use by %d models.", attachedModels),
			})
			return
		}

		result, err := db.ExecContext(r.Context(), `DELETE FROM providers WHERE id = $1`, id)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to delete provider"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			WriteError(w, http.StatusNotFound, "provider not found", "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleCreateAPIKey inserts a new API key row and returns the generated raw key once.
func HandleCreateAPIKey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		var payload createAPIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
			return
		}

		name := strings.TrimSpace(payload.Name)
		if name == "" || payload.RateLimitRPS < 0 {
			WriteError(w, http.StatusBadRequest, "name and non-negative rate_limit_rps are required", "invalid_request_error")
			return
		}
		if payload.Metadata == nil {
			payload.Metadata = map[string]string{}
		}
		metadataJSON, err := json.Marshal(payload.Metadata)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "metadata must be a string map", "invalid_request_error")
			return
		}

		allowedModelsParam := textArrayLiteral(payload.AllowedModels)

		rawKey, err := generateRawAPIKey()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to generate api key", "server_error")
			return
		}

		now := time.Now().UTC()
		id := uuid.NewString()
		keyHash := hashAPIKey(rawKey)
		_, err = db.ExecContext(
			r.Context(),
			`INSERT INTO api_keys (
				id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, allowed_models, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::text[], $9, $10)`,
			id,
			keyHash,
			name,
			payload.RateLimitRPS,
			true,
			payload.ExpiresAt,
			metadataJSON,
			allowedModelsParam,
			now,
			now,
		)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to create api key"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		WriteJSON(w, http.StatusCreated, createAPIKeyResponse{
			ID:            id,
			Name:          name,
			RateLimitRPS:  payload.RateLimitRPS,
			IsActive:      true,
			ExpiresAt:     payload.ExpiresAt,
			Metadata:      payload.Metadata,
			AllowedModels: payload.AllowedModels,
			APIKey:        rawKey,
		})
	}
}

// HandleUpdateAPIKey updates mutable API key fields by key ID.
func HandleUpdateAPIKey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "api key id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "api key id must be a valid UUID", "invalid_request_error")
			return
		}

		var payload updateAPIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
			return
		}

		setName := payload.Name != nil
		name := ""
		if setName {
			name = strings.TrimSpace(*payload.Name)
			if name == "" {
				WriteError(w, http.StatusBadRequest, "name cannot be empty", "invalid_request_error")
				return
			}
		}

		setRateLimit := payload.RateLimitRPS != nil
		rateLimit := 0
		if setRateLimit {
			rateLimit = *payload.RateLimitRPS
			if rateLimit < 0 {
				WriteError(w, http.StatusBadRequest, "rate_limit_rps must be non-negative", "invalid_request_error")
				return
			}
		}

		setExpiresAt := payload.ExpiresAt.Set
		expiresAt := payload.ExpiresAt.Value

		setMetadata := payload.Metadata != nil
		var metadataJSON []byte
		if setMetadata {
			var err error
			metadataJSON, err = json.Marshal(payload.Metadata)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "metadata must be a string map", "invalid_request_error")
				return
			}
		}

		setAllowedModels := payload.AllowedModels != nil
		allowedModels := textArrayLiteral(payload.AllowedModels)

		setIsActive := payload.IsActive != nil
		isActive := false
		if setIsActive {
			isActive = *payload.IsActive
		}

		if !setName && !setRateLimit && !setExpiresAt && !setMetadata && !setAllowedModels && !setIsActive {
			WriteError(w, http.StatusBadRequest, "at least one mutable field must be provided", "invalid_request_error")
			return
		}

		updated := llm.APIKey{}
		var metadataRaw []byte
		var allowedModelsRaw []byte
		err := db.QueryRowContext(
			r.Context(),
			`UPDATE api_keys
			 SET name = CASE WHEN $1 THEN $2 ELSE name END,
			     rate_limit_rps = CASE WHEN $3 THEN $4 ELSE rate_limit_rps END,
			     expires_at = CASE WHEN $5 THEN $6 ELSE expires_at END,
			     metadata = CASE WHEN $7 THEN $8::jsonb ELSE metadata END,
			     allowed_models = CASE WHEN $9 THEN $10::text[] ELSE allowed_models END,
			     is_active = CASE WHEN $11 THEN $12 ELSE is_active END,
			     updated_at = $13
			 WHERE id = $14
			 RETURNING id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, to_json(allowed_models), created_at, updated_at`,
			setName,
			name,
			setRateLimit,
			rateLimit,
			setExpiresAt,
			expiresAt,
			setMetadata,
			metadataJSON,
			setAllowedModels,
			allowedModels,
			setIsActive,
			isActive,
			time.Now().UTC(),
			id,
		).Scan(
			&updated.ID,
			&updated.KeyHash,
			&updated.Name,
			&updated.RateLimitRPS,
			&updated.IsActive,
			&updated.ExpiresAt,
			&metadataRaw,
			&allowedModelsRaw,
			&updated.CreatedAt,
			&updated.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteError(w, http.StatusNotFound, "api key not found", "invalid_request_error")
				return
			}
			status := http.StatusInternalServerError
			msg := "failed to update api key"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &updated.Metadata); err != nil {
				WriteError(w, http.StatusInternalServerError, "failed to parse api key metadata", "server_error")
				return
			}
		}
		if len(allowedModelsRaw) > 0 && string(allowedModelsRaw) != "null" {
			if err := json.Unmarshal(allowedModelsRaw, &updated.AllowedModels); err != nil {
				WriteError(w, http.StatusInternalServerError, "failed to parse api key allowed models", "server_error")
				return
			}
		}

		WriteJSON(w, http.StatusOK, updated)
	}
}

// HandleDeleteAPIKey hard-deletes an API key and related budget row by key ID.
func HandleDeleteAPIKey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "api key id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "api key id must be a valid UUID", "invalid_request_error")
			return
		}

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to delete api key", "invalid_request_error")
			return
		}
		defer tx.Rollback()

		result, err := tx.ExecContext(r.Context(), `DELETE FROM api_keys WHERE id = $1`, id)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to delete api key"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			WriteError(w, http.StatusNotFound, "api key not found", "invalid_request_error")
			return
		}

		if _, err := tx.ExecContext(r.Context(), `DELETE FROM api_key_budgets WHERE api_key_id = $1`, id); err != nil {
			status := http.StatusInternalServerError
			msg := "failed to delete api key budget"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		if err := tx.Commit(); err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to delete api key", "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleGetAPIKeyBudget returns the configured budget state for an API key.
func HandleGetAPIKeyBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "api key id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "api key id must be a valid UUID", "invalid_request_error")
			return
		}

		state, err := storage.GetBudget(r.Context(), db, id)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to fetch api key budget"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}
		if state == nil {
			WriteError(w, http.StatusNotFound, "api key budget not found", "invalid_request_error")
			return
		}

		WriteJSON(w, http.StatusOK, toBudgetResponse(state))
	}
}

// HandlePutAPIKeyBudget upserts the budget limit for an API key.
func HandlePutAPIKeyBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "api key id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "api key id must be a valid UUID", "invalid_request_error")
			return
		}

		var payload upsertBudgetRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
			return
		}
		if payload.BudgetUSD <= 0 {
			WriteError(w, http.StatusBadRequest, "budget_usd must be greater than 0", "invalid_request_error")
			return
		}
		if payload.AlertThresholdPct != nil {
			if *payload.AlertThresholdPct < 1 || *payload.AlertThresholdPct > 99 {
				WriteError(w, http.StatusBadRequest, "alert_threshold_pct must be between 1 and 99", "invalid_request_error")
				return
			}
		}

		var normalizedDuration *string
		if payload.BudgetDuration != nil {
			normalized, ok := normalizeBudgetDuration(*payload.BudgetDuration)
			if !ok {
				WriteError(w, http.StatusBadRequest, "budget_duration must be one of: daily, weekly, monthly", "invalid_request_error")
				return
			}
			normalizedDuration = &normalized
			if payload.BudgetResetAt == nil {
				WriteError(w, http.StatusBadRequest, "budget_reset_at is required when budget_duration is set", "invalid_request_error")
				return
			}
		}

		state, err := storage.UpsertBudget(
			r.Context(),
			db,
			id,
			payload.BudgetUSD,
			payload.AlertThresholdPct,
			normalizedDuration,
			payload.BudgetResetAt,
		)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to upsert api key budget"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		WriteJSON(w, http.StatusOK, toBudgetResponse(state))
	}
}

// HandleDeleteAPIKeyBudget removes a configured budget for an API key.
func HandleDeleteAPIKeyBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "api key id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "api key id must be a valid UUID", "invalid_request_error")
			return
		}

		if err := storage.DeleteBudget(r.Context(), db, id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteError(w, http.StatusNotFound, "api key budget not found", "invalid_request_error")
				return
			}
			status := http.StatusInternalServerError
			msg := "failed to delete api key budget"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleResetAPIKeyBudget resets an API key's spend to zero.
func HandleResetAPIKeyBudget(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			WriteError(w, http.StatusServiceUnavailable, "database unavailable", "service_unavailable")
			return
		}

		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			WriteError(w, http.StatusBadRequest, "api key id is required", "invalid_request_error")
			return
		}
		if _, err := uuid.Parse(id); err != nil {
			WriteError(w, http.StatusBadRequest, "api key id must be a valid UUID", "invalid_request_error")
			return
		}

		state, err := storage.ResetBudget(r.Context(), db, id)
		if err != nil {
			status := http.StatusInternalServerError
			msg := "failed to reset api key budget"
			if errors.Is(err, sql.ErrConnDone) {
				status = http.StatusServiceUnavailable
				msg = "database unavailable"
			}
			WriteError(w, status, msg, "invalid_request_error")
			return
		}
		if state == nil {
			WriteError(w, http.StatusNotFound, "api key budget not found", "invalid_request_error")
			return
		}

		WriteJSON(w, http.StatusOK, toBudgetResponse(state))
	}
}

func toBudgetResponse(state *storage.BudgetState) budgetResponse {
	var alertThreshold *int
	if state.AlertThresholdPct > 0 {
		alert := state.AlertThresholdPct
		alertThreshold = &alert
	}

	var budgetDuration *string
	if state.BudgetDuration != "" {
		duration := state.BudgetDuration
		budgetDuration = &duration
	}

	return budgetResponse{
		APIKeyID:          state.APIKeyID,
		BudgetUSD:         state.BudgetUSD,
		SpentUSD:          state.SpentUSD,
		RemainingUSD:      state.BudgetUSD - state.SpentUSD,
		AlertThresholdPct: alertThreshold,
		BudgetDuration:    budgetDuration,
		BudgetResetAt:     state.BudgetResetAt,
	}
}

func normalizeBudgetDuration(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "daily":
		return "daily", true
	case "weekly":
		return "weekly", true
	case "monthly":
		return "monthly", true
	default:
		return "", false
	}
}

func textArrayLiteral(values []string) *string {
	if values == nil {
		return nil
	}
	if len(values) == 0 {
		empty := "{}"
		return &empty
	}

	escaped := make([]string, 0, len(values))
	for _, v := range values {
		safe := strings.ReplaceAll(v, `\`, `\\`)
		safe = strings.ReplaceAll(safe, `"`, `\"`)
		escaped = append(escaped, `"`+safe+`"`)
	}
	literal := "{" + strings.Join(escaped, ",") + "}"
	return &literal
}

func encryptProviderCredential(key, plaintext []byte) (ciphertext []byte, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func generateRawAPIKey() (string, error) {
	const keyBytes = 24
	buf := make([]byte, keyBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "lgk_" + hex.EncodeToString(buf), nil
}

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return fmt.Sprintf("%x", sum[:])
}
