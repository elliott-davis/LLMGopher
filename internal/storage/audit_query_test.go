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
