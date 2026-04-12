package storage

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueryUsage_GroupByModel_AppliesFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

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

	got, err := QueryUsage(context.Background(), db, UsageQuery{
		GroupBy:  "model",
		From:     from,
		To:       to,
		APIKeyID: "key-001",
		Model:    "gpt-4o",
	})
	if err != nil {
		t.Fatalf("QueryUsage: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("summary count = %d, want 1", len(got))
	}
	if got[0].Group != "gpt-4o" {
		t.Fatalf("group = %q, want gpt-4o", got[0].Group)
	}
	if got[0].Errors != 12 {
		t.Fatalf("errors = %d, want 12", got[0].Errors)
	}
	if got[0].AvgLatencyMS != 834.0 {
		t.Fatalf("avg_latency_ms = %v, want 834", got[0].AvgLatencyMS)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestQueryUsage_InvalidGroupBy_ReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	_, err = QueryUsage(context.Background(), db, UsageQuery{GroupBy: "tenant"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !regexp.MustCompile(`invalid group_by`).MatchString(err.Error()) {
		t.Fatalf("error = %q, want invalid group_by", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestQueryDailyUsage_AppliesFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.January, 8, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`(?s)SELECT\s+DATE_TRUNC\('day', created_at\)::date::text AS day,.*FROM audit_log\s+WHERE created_at >= \$1 AND created_at <= \$2 AND api_key_id = \$3 AND model = \$4\s+GROUP BY DATE_TRUNC\('day', created_at\)::date\s+ORDER BY DATE_TRUNC\('day', created_at\)::date ASC`).
		WithArgs(from, to, "key-001", "gpt-4o").
		WillReturnRows(sqlmock.NewRows([]string{"day", "requests", "total_tokens", "cost_usd"}).
			AddRow("2025-01-01", int64(120), int64(45000), 1.02).
			AddRow("2025-01-02", int64(95), int64(38000), 0.89))

	got, err := QueryDailyUsage(context.Background(), db, from, to, "key-001", "gpt-4o")
	if err != nil {
		t.Fatalf("QueryDailyUsage: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("summary count = %d, want 2", len(got))
	}
	if got[0].Date != "2025-01-01" || got[0].Requests != 120 {
		t.Fatalf("first row = %+v, want date=2025-01-01 requests=120", got[0])
	}
	if got[1].CostUSD != 0.89 {
		t.Fatalf("second row cost = %v, want 0.89", got[1].CostUSD)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
