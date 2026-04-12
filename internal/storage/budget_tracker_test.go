package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ed007183/llmgopher/internal/storage"
)

func TestGetBudget_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT api_key_id, budget_usd, spent_usd FROM api_key_budgets WHERE api_key_id = $1`)).
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnRows(sqlmock.NewRows([]string{"api_key_id", "budget_usd", "spent_usd"}).AddRow(
			"11111111-1111-1111-1111-111111111111", 100.0, 23.45,
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
	if state.ResetAt != nil {
		t.Fatalf("reset_at = %v, want nil", state.ResetAt)
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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT api_key_id, budget_usd, spent_usd FROM api_key_budgets WHERE api_key_id = $1`)).
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

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO api_key_budgets (api_key_id, budget_usd, spent_usd, created_at, updated_at)
		 VALUES ($1, $2, 0, NOW(), NOW())
		 ON CONFLICT (api_key_id) DO UPDATE
		 SET budget_usd = EXCLUDED.budget_usd, updated_at = NOW()
		 RETURNING api_key_id, budget_usd, spent_usd`)).
		WithArgs("11111111-1111-1111-1111-111111111111", 200.0).
		WillReturnRows(sqlmock.NewRows([]string{"api_key_id", "budget_usd", "spent_usd"}).AddRow(
			"11111111-1111-1111-1111-111111111111", 200.0, 50.5,
		))

	state, err := storage.UpsertBudget(context.Background(), db, "11111111-1111-1111-1111-111111111111", 200.0)
	if err != nil {
		t.Fatalf("UpsertBudget returned error: %v", err)
	}
	if state == nil {
		t.Fatal("state is nil")
	}
	if state.BudgetUSD != 200.0 || state.SpentUSD != 50.5 {
		t.Fatalf("state = %+v, want budget=200 spent=50.5", state)
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
		 SET spent_usd = 0, updated_at = NOW()
		 WHERE api_key_id = $1
		 RETURNING api_key_id, budget_usd, spent_usd`)).
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
