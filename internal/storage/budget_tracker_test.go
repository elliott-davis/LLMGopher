package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ed007183/llmgopher/internal/storage"
)

func TestGetBudget_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	resetAt := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	lastAlertedAt := resetAt.Add(-2 * time.Hour)

	mock.ExpectQuery(`SELECT[\s\S]+FROM api_key_budgets[\s\S]+WHERE api_key_id = \$1`).
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnRows(sqlmock.NewRows([]string{
			"api_key_id",
			"budget_usd",
			"spent_usd",
			"alert_threshold_pct",
			"budget_duration",
			"budget_reset_at",
			"last_alerted_at",
		}).AddRow(
			"11111111-1111-1111-1111-111111111111",
			100.0,
			23.45,
			80,
			"monthly",
			resetAt,
			lastAlertedAt,
		))

	state, err := storage.GetBudget(context.Background(), db, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("GetBudget returned error: %v", err)
	}
	if state == nil {
		t.Fatal("state is nil")
	}
	if state.BudgetUSD != 100.0 || state.SpentUSD != 23.45 {
		t.Fatalf("state = %+v, want budget=100 spent=23.45", state)
	}
	if state.AlertThresholdPct != 80 {
		t.Fatalf("alert_threshold_pct = %d, want 80", state.AlertThresholdPct)
	}
	if state.BudgetDuration != "monthly" {
		t.Fatalf("budget_duration = %q, want monthly", state.BudgetDuration)
	}
	if state.BudgetResetAt == nil || !state.BudgetResetAt.Equal(resetAt) {
		t.Fatalf("budget_reset_at = %v, want %v", state.BudgetResetAt, resetAt)
	}
	if state.LastAlertedAt == nil || !state.LastAlertedAt.Equal(lastAlertedAt) {
		t.Fatalf("last_alerted_at = %v, want %v", state.LastAlertedAt, lastAlertedAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGetBudget_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT[\s\S]+FROM api_key_budgets[\s\S]+WHERE api_key_id = \$1`).
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnError(sql.ErrNoRows)

	state, err := storage.GetBudget(context.Background(), db, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("GetBudget returned error: %v", err)
	}
	if state != nil {
		t.Fatalf("state = %+v, want nil", state)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestUpsertBudget_ReturnsState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	threshold := 80
	duration := "monthly"
	resetAt := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	lastAlertedAt := resetAt.Add(-45 * time.Minute)

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
		WithArgs("11111111-1111-1111-1111-111111111111", 200.0, &threshold, &duration, &resetAt).
		WillReturnRows(sqlmock.NewRows([]string{
			"api_key_id",
			"budget_usd",
			"spent_usd",
			"alert_threshold_pct",
			"budget_duration",
			"budget_reset_at",
			"last_alerted_at",
		}).AddRow(
			"11111111-1111-1111-1111-111111111111",
			200.0,
			50.5,
			80,
			"monthly",
			resetAt,
			lastAlertedAt,
		))

	state, err := storage.UpsertBudget(
		context.Background(),
		db,
		"11111111-1111-1111-1111-111111111111",
		200.0,
		&threshold,
		&duration,
		&resetAt,
	)
	if err != nil {
		t.Fatalf("UpsertBudget returned error: %v", err)
	}
	if state == nil {
		t.Fatal("state is nil")
	}
	if state.BudgetUSD != 200.0 || state.SpentUSD != 50.5 {
		t.Fatalf("state = %+v, want budget=200 spent=50.5", state)
	}
	if state.AlertThresholdPct != 80 {
		t.Fatalf("alert_threshold_pct = %d, want 80", state.AlertThresholdPct)
	}
	if state.BudgetDuration != "monthly" {
		t.Fatalf("budget_duration = %q, want monthly", state.BudgetDuration)
	}
	if state.BudgetResetAt == nil || !state.BudgetResetAt.Equal(resetAt) {
		t.Fatalf("budget_reset_at = %v, want %v", state.BudgetResetAt, resetAt)
	}
	if state.LastAlertedAt == nil || !state.LastAlertedAt.Equal(lastAlertedAt) {
		t.Fatalf("last_alerted_at = %v, want %v", state.LastAlertedAt, lastAlertedAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestDeleteBudget_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM api_key_budgets WHERE api_key_id = $1`)).
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = storage.DeleteBudget(context.Background(), db, "11111111-1111-1111-1111-111111111111")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestResetBudget_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

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
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnError(sql.ErrNoRows)

	state, err := storage.ResetBudget(context.Background(), db, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("ResetBudget returned error: %v", err)
	}
	if state != nil {
		t.Fatalf("state = %+v, want nil", state)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMarkBudgetAlerted_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	alertedAt := time.Now().UTC()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE api_key_budgets SET last_alerted_at = $1, updated_at = NOW() WHERE api_key_id = $2`)).
		WithArgs(alertedAt, "11111111-1111-1111-1111-111111111111").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = storage.MarkBudgetAlerted(context.Background(), db, "11111111-1111-1111-1111-111111111111", alertedAt)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
