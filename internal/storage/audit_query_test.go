package storage

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueryAuditLog_AppliesFiltersAndPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.January, 31, 23, 59, 59, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM audit_log WHERE api_key_id = $1 AND model = $2 AND provider = $3 AND status_code >= 400 AND created_at >= $4 AND created_at <= $5`,
	)).
		WithArgs("key-001", "gpt-4o", "openai", from, to).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider,\s+prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at\s+FROM audit_log WHERE api_key_id = \$1 AND model = \$2 AND provider = \$3 AND status_code >= 400 AND created_at >= \$4 AND created_at <= \$5\s+ORDER BY created_at DESC, id DESC\s+LIMIT \$6 OFFSET \$7`).
		WithArgs("key-001", "gpt-4o", "openai", from, to, 50, 20).
		WillReturnRows(
			sqlmock.NewRows([]string{
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
				500,
				int64(823),
				false,
				"upstream timeout",
				time.Date(2025, time.January, 10, 12, 0, 0, 0, time.UTC),
			),
		)

	result, err := QueryAuditLog(context.Background(), db, AuditQuery{
		APIKeyID: "key-001",
		Model:    "gpt-4o",
		Provider: "openai",
		Status:   "error",
		From:     &from,
		To:       &to,
		Limit:    50,
		Offset:   20,
	})
	if err != nil {
		t.Fatalf("QueryAuditLog: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("total = %d, want 1", result.Total)
	}
	if len(result.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(result.Data))
	}
	if result.Data[0].ID != 1234 {
		t.Fatalf("id = %d, want 1234", result.Data[0].ID)
	}
	if got := result.Data[0].Latency.Milliseconds(); got != 823 {
		t.Fatalf("latency_ms = %d, want 823", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestQueryAuditLog_ActorFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM audit_log WHERE api_key_id = $1`,
	)).
		WithArgs("key-actor-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider,\s+prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at\s+FROM audit_log WHERE api_key_id = \$1\s+ORDER BY created_at DESC, id DESC\s+LIMIT \$2 OFFSET \$3`).
		WithArgs("key-actor-001", 100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).
			AddRow(int64(1), "r1", "key-actor-001", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()).
			AddRow(int64(2), "r2", "key-actor-001", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()))

	result, err := QueryAuditLog(context.Background(), db, AuditQuery{APIKeyID: "key-actor-001"})
	if err != nil {
		t.Fatalf("QueryAuditLog: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("total = %d, want 2", result.Total)
	}
	for _, row := range result.Data {
		if row.APIKeyID != "key-actor-001" {
			t.Errorf("unexpected api_key_id %q, want key-actor-001", row.APIKeyID)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryAuditLog_ActionFamilyFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM audit_log WHERE model IS NOT NULL`,
	)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider,\s+prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at\s+FROM audit_log WHERE model IS NOT NULL\s+ORDER BY created_at DESC, id DESC\s+LIMIT \$1 OFFSET \$2`).
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}))

	result, err := QueryAuditLog(context.Background(), db, AuditQuery{
		Action:      "request:",
		ActionExact: false,
	})
	if err != nil {
		t.Fatalf("QueryAuditLog: %v", err)
	}
	if result.Total != 5 {
		t.Fatalf("total = %d, want 5", result.Total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryAuditLog_ActionExactFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM audit_log WHERE model = $1`,
	)).
		WithArgs("claude-3-opus").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider,\s+prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at\s+FROM audit_log WHERE model = \$1\s+ORDER BY created_at DESC, id DESC\s+LIMIT \$2 OFFSET \$3`).
		WithArgs("claude-3-opus", 100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}))

	result, err := QueryAuditLog(context.Background(), db, AuditQuery{
		Action:      "request:claude-3-opus",
		ActionExact: true,
	})
	if err != nil {
		t.Fatalf("QueryAuditLog: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("total = %d, want 3", result.Total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryAuditLog_OutcomeFilters(t *testing.T) {
	tests := []struct {
		outcome         string
		expectedWhere   string
	}{
		{"success", "status_code < 400"},
		{"unauthorized", "(status_code = 401 OR status_code = 403)"},
		{"budget_denied", "(status_code = 429 AND error_message ILIKE '%budget%')"},
		{"rate_limited", "(status_code = 429 AND error_message NOT ILIKE '%budget%')"},
		{"client_error", "(status_code >= 400 AND status_code < 500 AND status_code NOT IN (401, 403, 429))"},
		{"failure", "status_code >= 500"},
	}

	for _, tc := range tests {
		t.Run(tc.outcome, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT count(*) FROM audit_log WHERE ` + tc.expectedWhere,
			)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

			mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider`).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "request_id", "api_key_id", "model", "provider",
					"prompt_tokens", "output_tokens", "total_tokens",
					"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
				}))

			_, err = QueryAuditLog(context.Background(), db, AuditQuery{Outcome: tc.outcome})
			if err != nil {
				t.Fatalf("QueryAuditLog: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestQueryAuditLog_DeterministicOrdering_IDTiebreaker(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sharedTime := time.Date(2025, time.January, 10, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// sqlmock returns rows in the order we add them; the ORDER BY is validated by
	// the regex below ensuring the clause appears in the query.
	mock.ExpectQuery(`ORDER BY created_at DESC, id DESC`).
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).
			AddRow(int64(30), "r3", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", sharedTime).
			AddRow(int64(20), "r2", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", sharedTime).
			AddRow(int64(10), "r1", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", sharedTime))

	result, err := QueryAuditLog(context.Background(), db, AuditQuery{})
	if err != nil {
		t.Fatalf("QueryAuditLog: %v", err)
	}
	if len(result.Data) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(result.Data))
	}
	// DB returns them newest-first by id (30, 20, 10) — verify the order is honoured
	wantIDs := []int64{30, 20, 10}
	for i, wantID := range wantIDs {
		if result.Data[i].ID != wantID {
			t.Errorf("row[%d].ID = %d, want %d", i, result.Data[i].ID, wantID)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryAuditLog_LargeOffsetPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Simulate a total of 65 rows; request page 3 (offset=40, limit=20).
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(65))

	mock.ExpectQuery(`ORDER BY created_at DESC, id DESC\s+LIMIT \$1 OFFSET \$2`).
		WithArgs(20, 40).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "request_id", "api_key_id", "model", "provider",
			"prompt_tokens", "output_tokens", "total_tokens",
			"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
		}).
			AddRow(int64(25), "r25", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()).
			AddRow(int64(24), "r24", "", "gpt-4o", "openai", 0, 0, 0, 0.0, 200, int64(10), false, "", time.Now()))

	result, err := QueryAuditLog(context.Background(), db, AuditQuery{Limit: 20, Offset: 40})
	if err != nil {
		t.Fatalf("QueryAuditLog: %v", err)
	}
	if result.Total != 65 {
		t.Fatalf("total = %d, want 65", result.Total)
	}
	if len(result.Data) != 2 {
		t.Fatalf("data length = %d, want 2", len(result.Data))
	}
	// Verify offset+len < total → has_more would be true at handler level
	if result.Total <= 40+len(result.Data) {
		t.Fatalf("expected more pages: total=%d offset+len=%d", result.Total, 40+len(result.Data))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryAuditLog_DefaultAndMaxLimit(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		offset    int
		wantLimit int
	}{
		{
			name:      "default limit when unset",
			limit:     0,
			offset:    0,
			wantLimit: 100,
		},
		{
			name:      "cap limit above max",
			limit:     5000,
			offset:    3,
			wantLimit: 1000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM audit_log`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

			mock.ExpectQuery(`SELECT id, request_id, api_key_id, model, provider,\s+prompt_tokens, output_tokens, total_tokens, cost_usd, status_code, latency_ms, streaming, error_message, created_at\s+FROM audit_log\s+ORDER BY created_at DESC, id DESC\s+LIMIT \$1 OFFSET \$2`).
				WithArgs(tc.wantLimit, tc.offset).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "request_id", "api_key_id", "model", "provider",
					"prompt_tokens", "output_tokens", "total_tokens",
					"cost_usd", "status_code", "latency_ms", "streaming", "error_message", "created_at",
				}))

			result, err := QueryAuditLog(context.Background(), db, AuditQuery{
				Limit:  tc.limit,
				Offset: tc.offset,
			})
			if err != nil {
				t.Fatalf("QueryAuditLog: %v", err)
			}
			if result.Total != 0 {
				t.Fatalf("total = %d, want 0", result.Total)
			}
			if len(result.Data) != 0 {
				t.Fatalf("data length = %d, want 0", len(result.Data))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL expectations: %v", err)
			}
		})
	}
}
