package api_test

import (
	"bytes"
	"database/sql"
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

func TestHandleGetAuditLog_NilDB_ReturnsServiceUnavailable(t *testing.T) {
	handler := api.HandleGetAuditLog(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleGetAuditLog_InvalidLimit_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?limit=oops", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetAuditLog_EmptyResult_ReturnsEmptyData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider,\s+prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at\s+FROM audit_log\s+ORDER BY created_at DESC, id DESC\s+LIMIT \$1 OFFSET \$2`).
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}))

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload struct {
		Data   []any `json:"data"`
		Total  int   `json:"total"`
		Limit  int   `json:"limit"`
		Offset int   `json:"offset"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data == nil {
		t.Fatal("data should be an empty array, got null")
	}
	if len(payload.Data) != 0 {
		t.Fatalf("data length = %d, want 0", len(payload.Data))
	}
	if payload.Total != 0 || payload.Limit != 100 || payload.Offset != 0 {
		t.Fatalf("pagination = (%d,%d,%d), want (0,100,0)", payload.Total, payload.Limit, payload.Offset)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleGetAuditLog_Success_CapsLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM audit_log WHERE api_key_id = $1 AND model = $2 AND provider = $3 AND status_code < 400 AND created_at >= $4 AND created_at <= $5`,
	)).
		WithArgs("key-001", "gpt-4o", "openai", from, to).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider,\s+prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at\s+FROM audit_log WHERE api_key_id = \$1 AND model = \$2 AND provider = \$3 AND status_code < 400 AND created_at >= \$4 AND created_at <= \$5\s+ORDER BY created_at DESC, id DESC\s+LIMIT \$6 OFFSET \$7`).
		WithArgs("key-001", "gpt-4o", "openai", from, to, 1000, 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).AddRow(
			int64(1234),
			"req-123",
			"key-001",
			"gpt-4o",
			"openai",
			100,
			50,
			150,
			0.00225,
			200,
			int64(823),
			false,
			"",
			time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC),
		))

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/admin/audit?api_key_id=key-001&model=gpt-4o&provider=openai&status=success&from=2025-01-01T00:00:00Z&to=2025-01-02T00:00:00Z&limit=5000&offset=10",
		nil,
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload struct {
		Data []struct {
			ID        int64  `json:"id"`
			RequestID string `json:"request_id"`
			LatencyMS int64  `json:"latency_ms"`
		} `json:"data"`
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Total != 1 || payload.Limit != 1000 || payload.Offset != 10 {
		t.Fatalf("pagination = (%d,%d,%d), want (1,1000,10)", payload.Total, payload.Limit, payload.Offset)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(payload.Data))
	}
	if payload.Data[0].ID != 1234 || payload.Data[0].RequestID != "req-123" || payload.Data[0].LatencyMS != 823 {
		t.Fatalf("entry = %+v, want id=1234 request_id=req-123 latency_ms=823", payload.Data[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleGetUsage_NilDB_ReturnsServiceUnavailable(t *testing.T) {
	handler := api.HandleGetUsage(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage?group_by=model", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleGetUsage_MissingGroupBy_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetUsage(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetUsage_InvalidGroupBy_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetUsage(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/usage?group_by=team", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetUsage_FromAfterTo_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetUsage(db)
	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/admin/usage?group_by=model&from=2025-01-02T00:00:00Z&to=2025-01-01T00:00:00Z",
		nil,
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetUsage_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetUsage(db)
	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`(?s)SELECT\s+model AS grp,.*FROM audit_log\s+WHERE created_at >= \$1 AND created_at <= \$2 AND api_key_id = \$3 AND model = \$4\s+GROUP BY model\s+ORDER BY cost_usd DESC, grp ASC`).
		WithArgs(from, to, "key-001", "gpt-4o").
		WillReturnRows(sqlmock.NewRows([]string{
			"grp",
			"requests",
			"prompt_tokens",
			"completion_tokens",
			"total_tokens",
			"cost_usd",
			"errors",
			"avg_latency_ms",
		}).AddRow("gpt-4o", int64(1240), int64(450000), int64(220000), int64(670000), 15.23, int64(12), 834.0))

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/admin/usage?group_by=model&api_key_id=key-001&model=gpt-4o&from=2025-01-01T00:00:00Z&to=2025-02-01T00:00:00Z",
		nil,
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload struct {
		GroupBy string `json:"group_by"`
		From    string `json:"from"`
		To      string `json:"to"`
		Data    []struct {
			Group        string  `json:"group"`
			Requests     int64   `json:"requests"`
			Errors       int64   `json:"errors"`
			AvgLatencyMS float64 `json:"avg_latency_ms"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.GroupBy != "model" {
		t.Fatalf("group_by = %q, want model", payload.GroupBy)
	}
	if payload.From != "2025-01-01T00:00:00Z" || payload.To != "2025-02-01T00:00:00Z" {
		t.Fatalf("window = (%s,%s), want (2025-01-01T00:00:00Z,2025-02-01T00:00:00Z)", payload.From, payload.To)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(payload.Data))
	}
	if payload.Data[0].Group != "gpt-4o" || payload.Data[0].Requests != 1240 || payload.Data[0].Errors != 12 {
		t.Fatalf("summary = %+v, want group=gpt-4o requests=1240 errors=12", payload.Data[0])
	}
	if payload.Data[0].AvgLatencyMS != 834.0 {
		t.Fatalf("avg_latency_ms = %v, want 834", payload.Data[0].AvgLatencyMS)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleGetDailyUsage_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetDailyUsage(db)
	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.January, 8, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`(?s)SELECT\s+DATE_TRUNC\('day', created_at\)::date::text AS day,.*FROM audit_log\s+WHERE created_at >= \$1 AND created_at <= \$2 AND api_key_id = \$3 AND model = \$4\s+GROUP BY DATE_TRUNC\('day', created_at\)::date\s+ORDER BY DATE_TRUNC\('day', created_at\)::date ASC`).
		WithArgs(from, to, "key-001", "gpt-4o").
		WillReturnRows(sqlmock.NewRows([]string{"day", "requests", "total_tokens", "cost_usd"}).
			AddRow("2025-01-01", int64(120), int64(45000), 1.02).
			AddRow("2025-01-02", int64(95), int64(38000), 0.89))

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/admin/usage/daily?api_key_id=key-001&model=gpt-4o&from=2025-01-01T00:00:00Z&to=2025-01-08T00:00:00Z",
		nil,
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload struct {
		From string `json:"from"`
		To   string `json:"to"`
		Data []struct {
			Date        string  `json:"date"`
			Requests    int64   `json:"requests"`
			TotalTokens int64   `json:"total_tokens"`
			CostUSD     float64 `json:"cost_usd"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.From != "2025-01-01T00:00:00Z" || payload.To != "2025-01-08T00:00:00Z" {
		t.Fatalf("window = (%s,%s), want (2025-01-01T00:00:00Z,2025-01-08T00:00:00Z)", payload.From, payload.To)
	}
	if len(payload.Data) != 2 {
		t.Fatalf("data length = %d, want 2", len(payload.Data))
	}
	if payload.Data[0].Date != "2025-01-01" || payload.Data[0].Requests != 120 || payload.Data[0].TotalTokens != 45000 {
		t.Fatalf("first row = %+v, want date=2025-01-01 requests=120 total_tokens=45000", payload.Data[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
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

func TestHandleCreateModel_SuccessWithRateLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleCreateModel(db)
	providerID := "99999999-9999-9999-9999-999999999999"

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO models (id, alias, name, provider_id, context_window, rate_limit_rps, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)).
		WithArgs(
			sqlmock.AnyArg(),
			"gpt-4o",
			"gpt-4o",
			providerID,
			128000,
			25,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/models",
		bytes.NewBufferString(`{"alias":"gpt-4o","name":"gpt-4o","provider_id":"99999999-9999-9999-9999-999999999999","context_window":128000,"rate_limit_rps":25}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleUpdateModel_SuccessWithRateLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleUpdateModel(db)
	modelID := "88888888-8888-8888-8888-888888888888"
	providerID := "77777777-7777-7777-7777-777777777777"

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE models SET alias = $1, name = $2, provider_id = $3, context_window = $4, rate_limit_rps = $5, updated_at = $6 WHERE id = $7`)).
		WithArgs(
			"gpt-4.1",
			"gpt-4.1",
			providerID,
			128000,
			10,
			sqlmock.AnyArg(),
			modelID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/models/"+modelID,
		bytes.NewBufferString(`{"alias":"gpt-4.1","name":"gpt-4.1","provider_id":"77777777-7777-7777-7777-777777777777","context_window":128000,"rate_limit_rps":10}`),
	)
	req.SetPathValue("id", modelID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
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

func TestHandleCreateProvider_BedrockMissingCredentialToken_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleCreateProvider(db, bytes.Repeat([]byte("k"), 32))
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/providers",
		bytes.NewBufferString(`{"name":"Bedrock","base_url":"us-east-1","auth_type":"aws_bedrock"}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateProvider_BedrockInvalidCredentialToken_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleCreateProvider(db, bytes.Repeat([]byte("k"), 32))
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/providers",
		bytes.NewBufferString(`{"name":"Bedrock","base_url":"us-east-1","auth_type":"aws_bedrock","credential_token":"not-json"}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
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

func TestHandleCreateAPIKey_WithEnhancements_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleCreateAPIKey(db)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO api_keys (
				id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, allowed_models, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::text[], $9, $10)`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"team-a",
			250,
			true,
			sqlmock.AnyArg(),
			[]byte(`{"tenant":"acme"}`),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/admin/keys",
		bytes.NewBufferString(`{
			"name":"team-a",
			"rate_limit_rps":250,
			"expires_at":"2027-01-01T12:00:00Z",
			"metadata":{"tenant":"acme"},
			"allowed_models":["gpt-4o"]
		}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["name"] != "team-a" {
		t.Fatalf("name = %v, want team-a", payload["name"])
	}
	if payload["rate_limit_rps"] != float64(250) {
		t.Fatalf("rate_limit_rps = %v, want 250", payload["rate_limit_rps"])
	}
	if payload["api_key"] == "" {
		t.Fatal("api_key should be present")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleUpdateAPIKey_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleUpdateAPIKey(db)
	keyID := "44444444-4444-4444-4444-444444444444"
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE api_keys
			 SET name = CASE WHEN $1 THEN $2 ELSE name END,
			     rate_limit_rps = CASE WHEN $3 THEN $4 ELSE rate_limit_rps END,
			     expires_at = CASE WHEN $5 THEN $6 ELSE expires_at END,
			     metadata = CASE WHEN $7 THEN $8::jsonb ELSE metadata END,
			     allowed_models = CASE WHEN $9 THEN $10::text[] ELSE allowed_models END,
			     is_active = CASE WHEN $11 THEN $12 ELSE is_active END,
			     updated_at = $13
			 WHERE id = $14
			 RETURNING id, key_hash, name, rate_limit_rps, is_active, expires_at, metadata, to_json(allowed_models), created_at, updated_at`)).
		WithArgs(
			true,
			"updated-key",
			true,
			500,
			true,
			nil,
			true,
			[]byte(`{"cost_center":"ml"}`),
			true,
			sqlmock.AnyArg(),
			true,
			false,
			sqlmock.AnyArg(),
			keyID,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "key_hash", "name", "rate_limit_rps", "is_active", "expires_at", "metadata", "allowed_models", "created_at", "updated_at"}).
				AddRow(keyID, "hash", "updated-key", 500, false, nil, []byte(`{"cost_center":"ml"}`), []byte(`["gpt-4o"]`), now, now),
		)

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/keys/"+keyID,
		bytes.NewBufferString(`{
			"name":"updated-key",
			"rate_limit_rps":500,
			"expires_at":null,
			"metadata":{"cost_center":"ml"},
			"allowed_models":["gpt-4o"],
			"is_active":false
		}`),
	)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["name"] != "updated-key" {
		t.Fatalf("name = %v, want updated-key", payload["name"])
	}
	if payload["is_active"] != false {
		t.Fatalf("is_active = %v, want false", payload["is_active"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleDeleteAPIKey_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleDeleteAPIKey(db)
	keyID := "55555555-5555-5555-5555-555555555555"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM api_keys WHERE id = $1`)).
		WithArgs(keyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM api_key_budgets WHERE api_key_id = $1`)).
		WithArgs(keyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/keys/"+keyID, nil)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleGetAPIKeyBudget_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	mock.ExpectQuery(`SELECT[\s\S]+FROM api_key_budgets[\s\S]+WHERE api_key_id = \$1`).
		WithArgs(keyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"api_key_id",
			"budget_usd",
			"spent_usd",
			"alert_threshold_pct",
			"budget_duration",
			"budget_reset_at",
			"last_alerted_at",
		}).AddRow(
			keyID, 100.0, 23.45, 0, "", nil, nil,
		))

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/keys/"+keyID+"/budget", nil)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["api_key_id"] != keyID {
		t.Fatalf("api_key_id = %v, want %s", payload["api_key_id"], keyID)
	}
	if payload["budget_usd"] != float64(100) {
		t.Fatalf("budget_usd = %v, want 100", payload["budget_usd"])
	}
	if payload["spent_usd"] != 23.45 {
		t.Fatalf("spent_usd = %v, want 23.45", payload["spent_usd"])
	}
	if payload["remaining_usd"] != 76.55 {
		t.Fatalf("remaining_usd = %v, want 76.55", payload["remaining_usd"])
	}
	if payload["alert_threshold_pct"] != nil {
		t.Fatalf("alert_threshold_pct = %v, want nil", payload["alert_threshold_pct"])
	}
	if payload["budget_duration"] != nil {
		t.Fatalf("budget_duration = %v, want nil", payload["budget_duration"])
	}
	if payload["budget_reset_at"] != nil {
		t.Fatalf("budget_reset_at = %v, want nil", payload["budget_reset_at"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleGetAPIKeyBudget_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	mock.ExpectQuery(`SELECT[\s\S]+FROM api_key_budgets[\s\S]+WHERE api_key_id = \$1`).
		WithArgs(keyID).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/keys/"+keyID+"/budget", nil)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandlePutAPIKeyBudget_InvalidBudget_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandlePutAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/keys/"+keyID+"/budget",
		bytes.NewBufferString(`{"budget_usd":0}`),
	)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePutAPIKeyBudget_InvalidAlertThreshold_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandlePutAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/keys/"+keyID+"/budget",
		bytes.NewBufferString(`{"budget_usd":10,"alert_threshold_pct":100}`),
	)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePutAPIKeyBudget_InvalidBudgetDuration_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandlePutAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/keys/"+keyID+"/budget",
		bytes.NewBufferString(`{"budget_usd":10,"budget_duration":"yearly","budget_reset_at":"2026-01-01T00:00:00Z"}`),
	)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePutAPIKeyBudget_DurationRequiresResetAt_ReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandlePutAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/keys/"+keyID+"/budget",
		bytes.NewBufferString(`{"budget_usd":10,"budget_duration":"monthly"}`),
	)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePutAPIKeyBudget_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandlePutAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"
	resetAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	lastAlertedAt := time.Date(2026, 5, 31, 23, 0, 0, 0, time.UTC)
	threshold := 80
	duration := "monthly"

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO api_key_budgets (
			api_key_id,
			budget_usd,
			spent_usd,
			alert_threshold_pct,
			budget_duration,
			budget_reset_at,
			last_alerted_at,
			created_at,
			updated_at
		)
		 VALUES ($1, $2, 0, $3, $4, $5, NULL, NOW(), NOW())
		 ON CONFLICT (api_key_id) DO UPDATE
		 SET budget_usd = EXCLUDED.budget_usd,
		     alert_threshold_pct = EXCLUDED.alert_threshold_pct,
		     budget_duration = EXCLUDED.budget_duration,
		     budget_reset_at = EXCLUDED.budget_reset_at,
		     updated_at = NOW()
		 RETURNING
		    api_key_id,
		    budget_usd,
		    spent_usd,
		    COALESCE(alert_threshold_pct, 0),
		    COALESCE(budget_duration, ''),
		    budget_reset_at,
		    last_alerted_at`)).
		WithArgs(keyID, 250.0, &threshold, &duration, &resetAt).
		WillReturnRows(sqlmock.NewRows([]string{
			"api_key_id",
			"budget_usd",
			"spent_usd",
			"alert_threshold_pct",
			"budget_duration",
			"budget_reset_at",
			"last_alerted_at",
		}).AddRow(
			keyID, 250.0, 120.5, 80, "monthly", resetAt, lastAlertedAt,
		))

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/admin/keys/"+keyID+"/budget",
		bytes.NewBufferString(`{"budget_usd":250,"alert_threshold_pct":80,"budget_duration":"monthly","budget_reset_at":"2026-06-01T00:00:00Z"}`),
	)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["budget_usd"] != float64(250) {
		t.Fatalf("budget_usd = %v, want 250", payload["budget_usd"])
	}
	if payload["spent_usd"] != 120.5 {
		t.Fatalf("spent_usd = %v, want 120.5", payload["spent_usd"])
	}
	if payload["remaining_usd"] != 129.5 {
		t.Fatalf("remaining_usd = %v, want 129.5", payload["remaining_usd"])
	}
	if payload["alert_threshold_pct"] != float64(80) {
		t.Fatalf("alert_threshold_pct = %v, want 80", payload["alert_threshold_pct"])
	}
	if payload["budget_duration"] != "monthly" {
		t.Fatalf("budget_duration = %v, want monthly", payload["budget_duration"])
	}
	if payload["budget_reset_at"] != "2026-06-01T00:00:00Z" {
		t.Fatalf("budget_reset_at = %v, want 2026-06-01T00:00:00Z", payload["budget_reset_at"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleDeleteAPIKeyBudget_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleDeleteAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM api_key_budgets WHERE api_key_id = $1`)).
		WithArgs(keyID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/keys/"+keyID+"/budget", nil)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleDeleteAPIKeyBudget_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleDeleteAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM api_key_budgets WHERE api_key_id = $1`)).
		WithArgs(keyID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/keys/"+keyID+"/budget", nil)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleResetAPIKeyBudget_SetsSpentToZero(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleResetAPIKeyBudget(db)
	keyID := "66666666-6666-6666-6666-666666666666"

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE api_key_budgets
		 SET spent_usd = 0, last_alerted_at = NULL, updated_at = NOW()
		 WHERE api_key_id = $1
		 RETURNING
		    api_key_id,
		    budget_usd,
		    spent_usd,
		    COALESCE(alert_threshold_pct, 0),
		    COALESCE(budget_duration, ''),
		    budget_reset_at,
		    last_alerted_at`)).
		WithArgs(keyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"api_key_id",
			"budget_usd",
			"spent_usd",
			"alert_threshold_pct",
			"budget_duration",
			"budget_reset_at",
			"last_alerted_at",
		}).AddRow(
			keyID, 250.0, 0.0, 0, "", nil, nil,
		))

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/keys/"+keyID+"/budget/reset", nil)
	req.SetPathValue("id", keyID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["spent_usd"] != float64(0) || payload["remaining_usd"] != float64(250) {
		t.Fatalf("payload = %+v, want spent_usd=0 remaining_usd=250", payload)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
