package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ed007183/llmgopher/internal/api"
	"github.com/ed007183/llmgopher/internal/validation"
)

func TestHandleGetKeys_EmptyCache_ReturnsEmptyArray(t *testing.T) {
	handler := api.HandleGetKeys(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/keys", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body []any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("body length = %d, want 0", len(body))
	}
}

func TestHandleGetModels_EmptyCache_ReturnsEmptyArray(t *testing.T) {
	handler := api.HandleGetModels(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/models", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body []any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("body length = %d, want 0", len(body))
	}
}

func TestHandleGetProviders_EmptyCache_ReturnsEmptyArray(t *testing.T) {
	handler := api.HandleGetProviders(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/providers", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body []any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("body length = %d, want 0", len(body))
	}
}

func TestHandleCreateModel_NilDB_ReturnsServiceUnavailable(t *testing.T) {
	handler := api.HandleCreateModel(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/models",
		bytes.NewBufferString(`{"alias":"gpt-4o","name":"gpt-4o","provider_id":"p1","context_window":128000}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleCreateProvider_NilDB_ReturnsServiceUnavailable(t *testing.T) {
	handler := api.HandleCreateProvider(nil, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/providers",
		bytes.NewBufferString(`{"name":"OpenAI Prod","base_url":"https://api.openai.com/v1","auth_type":"bearer"}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleCreateAPIKey_NilDB_ReturnsServiceUnavailable(t *testing.T) {
	handler := api.HandleCreateAPIKey(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/keys",
		bytes.NewBufferString(`{"name":"test-key","rate_limit_rps":100}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleCreateAPIKey_InvalidPayload_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleCreateAPIKey(db)
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/keys",
		bytes.NewBufferString(`{"name":"","rate_limit_rps":-1}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleValidateCredential_NilValidator_ReturnsServiceUnavailable(t *testing.T) {
	handler := api.HandleValidateCredential(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/credentials/validate",
		bytes.NewBufferString(`{"provider":"openai","apiKey":"sk-test"}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleValidateCredential_ValidOpenAI_ReturnsOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	validator := validation.NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
	validator.SetOpenAIBaseURLForTest(srv.URL)
	handler := api.HandleValidateCredential(validator)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/credentials/validate",
		bytes.NewBufferString(`{"provider":"openai","apiKey":"sk-test"}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["valid"] != true {
		t.Fatalf("valid = %v, want true", payload["valid"])
	}
}

func TestHandleValidateCredential_InvalidKey_ReturnsUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_api_key"}`))
	}))
	defer srv.Close()

	validator := validation.NewCredentialValidator(&http.Client{Timeout: 2 * time.Second})
	validator.SetOpenAIBaseURLForTest(srv.URL)
	handler := api.HandleValidateCredential(validator)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/credentials/validate",
		bytes.NewBufferString(`{"provider":"openai","apiKey":"bad-key"}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["code"] != validation.ErrCodeInvalidAPIKey {
		t.Fatalf("code = %v, want %q", payload["code"], validation.ErrCodeInvalidAPIKey)
	}
}

func TestHandleUpdateProvider_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleUpdateProvider(db, bytes.Repeat([]byte("k"), 32))
	providerID := "11111111-1111-1111-1111-111111111111"

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE providers
			 SET name = $1,
			     base_url = $2,
			     auth_type = $3,
			     updated_at = $4,
			     credential_file_name = CASE WHEN $5 THEN $6 ELSE credential_file_name END,
			     credential_ciphertext = CASE WHEN $5 THEN $7 ELSE credential_ciphertext END,
			     credential_nonce = CASE WHEN $5 THEN $8 ELSE credential_nonce END,
			     has_credentials = CASE WHEN $5 THEN $9 ELSE has_credentials END
			 WHERE id = $10`)).
		WithArgs(
			"OpenAI Prod",
			"https://api.openai.com/v1",
			"bearer",
			sqlmock.AnyArg(),
			false,
			"",
			[]byte(nil),
			[]byte(nil),
			false,
			providerID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/providers/"+providerID,
		bytes.NewBufferString(`{"name":"OpenAI Prod","base_url":"https://api.openai.com/v1","auth_type":"bearer"}`),
	)
	req.SetPathValue("id", providerID)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleDeleteProvider_InUse_ReturnsConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleDeleteProvider(db)
	providerID := "22222222-2222-2222-2222-222222222222"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM models WHERE provider_id = $1`)).
		WithArgs(providerID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/providers/"+providerID, nil)
	req.SetPathValue("id", providerID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
	}

	var payload map[string]string
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	wantMessage := "Cannot delete provider. It is currently in use by 3 models."
	if payload["error"] != wantMessage {
		t.Fatalf("error message = %q, want %q", payload["error"], wantMessage)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleDeleteProvider_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleDeleteProvider(db)
	providerID := "33333333-3333-3333-3333-333333333333"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM models WHERE provider_id = $1`)).
		WithArgs(providerID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM providers WHERE id = $1`)).
		WithArgs(providerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/providers/"+providerID, nil)
	req.SetPathValue("id", providerID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
