package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ed007183/llmgopher/internal/api"
)

// emptyAuditRows returns sqlmock rows with the audit SELECT columns but no rows.
func emptyAuditRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "request_id", "api_key_id", "model", "provider",
		"prompt_tokens", "output_tokens", "total_tokens",
		"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
	})
}

// --- T012: new parameter parsing ---

func TestHandleGetAuditLog_ActorParam_MapsToAPIKeyID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log WHERE api_key_id = $1`)).
		WithArgs("key-actor-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT id, request_id`).
		WillReturnRows(emptyAuditRows())

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?actor=key-actor-001", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleGetAuditLog_ActorAndAPIKeyID_Returns400AmbiguousActor(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?actor=key-001&api_key_id=key-002", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
			Type string `json:"type"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "ambiguous_actor" {
		t.Errorf("code = %q, want ambiguous_actor", body.Error.Code)
	}
	if body.Error.Type != "invalid_request_error" {
		t.Errorf("type = %q, want invalid_request_error", body.Error.Type)
	}
}

func TestHandleGetAuditLog_OutcomeWinsOverStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// outcome=failure → status_code >= 500; legacy status=success should be ignored
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log WHERE status_code >= 500`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT id, request_id`).
		WillReturnRows(emptyAuditRows())

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?outcome=failure&status=success", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

// --- T013: validation failure paths ---

func TestHandleGetAuditLog_InvalidOutcome_Returns400(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?outcome=bogus", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	var body struct {
		Error struct{ Type string `json:"type"` } `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Type != "invalid_request_error" {
		t.Errorf("type = %q, want invalid_request_error", body.Error.Type)
	}
}

func TestHandleGetAuditLog_InvalidAction_Returns400(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?action=admin:keys.create", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestHandleGetAuditLog_MalformedFrom_Returns400(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?from=not-a-date", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestHandleGetAuditLog_FromAfterTo_Returns400(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?from=2025-02-01T00:00:00Z&to=2025-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestHandleGetAuditLog_NegativeOffset_Returns400(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?offset=-5", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// --- T014: pagination metadata ---

func TestHandleGetAuditLog_PaginationMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// 65 total, page 3 of 20 → offset=40, limit=20, page=3, has_more=true
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(65))
	mock.ExpectQuery(`ORDER BY created_at DESC, id DESC\s+LIMIT \$1 OFFSET \$2`).
		WithArgs(20, 40).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).
			AddRow(int64(1), "r1", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()).
			AddRow(int64(2), "r2", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()))

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?limit=20&offset=40", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	var payload struct {
		Total   int  `json:"total"`
		Limit   int  `json:"limit"`
		Offset  int  `json:"offset"`
		Page    int  `json:"page"`
		HasMore bool `json:"has_more"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Total != 65 {
		t.Errorf("total = %d, want 65", payload.Total)
	}
	if payload.Limit != 20 {
		t.Errorf("limit = %d, want 20", payload.Limit)
	}
	if payload.Offset != 40 {
		t.Errorf("offset = %d, want 40", payload.Offset)
	}
	wantPage := 40/20 + 1 // = 3
	if payload.Page != wantPage {
		t.Errorf("page = %d, want %d", payload.Page, wantPage)
	}
	// offset(40) + len(data)(2) = 42 < total(65) → has_more = true
	if !payload.HasMore {
		t.Errorf("has_more = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestHandleGetAuditLog_LastPage_HasMoreFalse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// 22 total, offset=20, limit=20 → 2 rows on this page; offset+2 = 22 = total → has_more = false
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(22))
	mock.ExpectQuery(`ORDER BY created_at DESC, id DESC`).
		WithArgs(20, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).
			AddRow(int64(1), "r1", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()).
			AddRow(int64(2), "r2", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()))

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit?limit=20&offset=20", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var payload struct {
		HasMore bool `json:"has_more"`
		Page    int  `json:"page"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.HasMore {
		t.Error("has_more = true on last page, want false")
	}
	if payload.Page != 2 {
		t.Errorf("page = %d, want 2", payload.Page)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

// --- T021: redaction integration ---

func TestHandleGetAuditLog_RedactsErrorMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	secretMsg := "upstream call failed: Bearer eyJhbGciOiJSUzI1NiJ9.payload.sig rejected"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, request_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).AddRow(int64(1), "r1", "key-001", "gpt-4o", "openai", 0, 0, 0, 0.0, 401, int64(10), false, secretMsg, time.Now()))

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !contains(body, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in response body, got: %s", body)
	}
	// The raw token must not appear
	if contains(body, "eyJhbGciOiJSUzI1NiJ9") {
		t.Errorf("bearer token still present in response body")
	}
}

// --- T022: missing-reference handler test ---

func TestHandleGetAuditLog_MissingActorID_EmitsReferenceSummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// Row with empty api_key_id
	mock.ExpectQuery(`SELECT id, request_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).AddRow(int64(1), "r1", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()))

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	var payload struct {
		Data []struct {
			ReferenceSummary []struct {
				Field string `json:"field"`
				State string `json:"state"`
			} `json:"reference_summary"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(payload.Data))
	}
	refs := payload.Data[0].ReferenceSummary
	if len(refs) == 0 {
		t.Fatal("expected reference_summary to be present for row with empty api_key_id")
	}
	found := false
	for _, r := range refs {
		if r.Field == "actor_id" && r.State == "missing" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reference_summary with field=actor_id state=missing, got %+v", refs)
	}
}

func TestHandleGetAuditLog_IntactReferences_NoReferenceSummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, request_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).AddRow(int64(1), "r1", "key-001", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()))

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// reference_summary must be omitted (not null) when all references are intact
	body := w.Body.String()
	if contains(body, "reference_summary") {
		t.Errorf("reference_summary should be omitted for intact row; body: %s", body)
	}
}

func TestHandleGetAuditLog_NewFieldsInResponse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, request_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).AddRow(int64(42), "r42", "key-001", "gpt-4o", "openai", 10, 20, 30, 0.002, 200, int64(100), false, "", time.Now()))

	handler := api.HandleGetAuditLog(db)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload struct {
		Data []struct {
			ActorID string `json:"actor_id"`
			APIKeyID string `json:"api_key_id"`
			Action  string `json:"action"`
			Outcome string `json:"outcome"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 row")
	}
	row := payload.Data[0]
	if row.ActorID != "key-001" {
		t.Errorf("actor_id = %q, want key-001", row.ActorID)
	}
	if row.APIKeyID != "key-001" {
		t.Errorf("api_key_id = %q, want key-001 (backward compat)", row.APIKeyID)
	}
	if row.Action != "request:gpt-4o" {
		t.Errorf("action = %q, want request:gpt-4o", row.Action)
	}
	if row.Outcome != "success" {
		t.Errorf("outcome = %q, want success", row.Outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
