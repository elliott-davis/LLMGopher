package storage

import (
	"context"
	"io"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestResetDueBudgets_ReturnsResetKeyIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE api_key_budgets
		 SET
		     spent_usd = 0,
		     budget_reset_at = CASE
		         WHEN budget_duration = 'daily'   THEN budget_reset_at + INTERVAL '1 day'
		         WHEN budget_duration = 'weekly'  THEN budget_reset_at + INTERVAL '1 week'
		         WHEN budget_duration = 'monthly' THEN budget_reset_at + INTERVAL '1 month'
		         ELSE budget_reset_at
		     END,
		     last_alerted_at = NULL,
		     updated_at = NOW()
		 WHERE budget_reset_at IS NOT NULL
		   AND budget_duration IS NOT NULL
		   AND budget_reset_at <= NOW()
		 RETURNING api_key_id`)).
		WillReturnRows(sqlmock.NewRows([]string{"api_key_id"}).
			AddRow("key-1").
			AddRow("key-2"))

	keyIDs, err := resetDueBudgets(context.Background(), db)
	if err != nil {
		t.Fatalf("resetDueBudgets returned error: %v", err)
	}
	if len(keyIDs) != 2 || keyIDs[0] != "key-1" || keyIDs[1] != "key-2" {
		t.Fatalf("keyIDs = %+v, want [key-1 key-2]", keyIDs)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestStartBudgetResetWorker_StopsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		StartBudgetResetWorker(ctx, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not stop after context cancellation")
	}
}
