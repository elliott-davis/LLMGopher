package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// PgBudgetTracker manages per-key spend limits backed by PostgreSQL.
// Budget checks use SELECT ... FOR UPDATE to prevent race conditions
// on concurrent deductions.
type PgBudgetTracker struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPgBudgetTracker(db *sql.DB, logger *slog.Logger) *PgBudgetTracker {
	return &PgBudgetTracker{db: db, logger: logger}
}

// BudgetState is the persisted budget configuration and current spend for an API key.
type BudgetState struct {
	APIKeyID  string
	BudgetUSD float64
	SpentUSD  float64
	ResetAt   *time.Time
}

func (bt *PgBudgetTracker) RemainingBudget(ctx context.Context, apiKeyID string) (float64, error) {
	var budget, spent float64
	err := bt.db.QueryRowContext(ctx,
		`SELECT budget_usd, spent_usd FROM api_key_budgets WHERE api_key_id = $1`,
		apiKeyID,
	).Scan(&budget, &spent)

	if err == sql.ErrNoRows {
		// No budget row means unlimited.
		return 1e9, nil
	}
	if err != nil {
		return 0, fmt.Errorf("query budget: %w", err)
	}
	remaining := budget - spent
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// GetBudget returns the current budget state for a key.
// A missing row returns (nil, nil).
func GetBudget(ctx context.Context, db *sql.DB, keyID string) (*BudgetState, error) {
	var state BudgetState
	err := db.QueryRowContext(
		ctx,
		`SELECT api_key_id, budget_usd, spent_usd FROM api_key_budgets WHERE api_key_id = $1`,
		keyID,
	).Scan(&state.APIKeyID, &state.BudgetUSD, &state.SpentUSD)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get budget: %w", err)
	}

	return &state, nil
}

// UpsertBudget creates or updates a budget row while preserving existing spent_usd.
func UpsertBudget(ctx context.Context, db *sql.DB, keyID string, budgetUSD float64) (*BudgetState, error) {
	var state BudgetState
	err := db.QueryRowContext(
		ctx,
		`INSERT INTO api_key_budgets (api_key_id, budget_usd, spent_usd, created_at, updated_at)
		 VALUES ($1, $2, 0, NOW(), NOW())
		 ON CONFLICT (api_key_id) DO UPDATE
		 SET budget_usd = EXCLUDED.budget_usd, updated_at = NOW()
		 RETURNING api_key_id, budget_usd, spent_usd`,
		keyID,
		budgetUSD,
	).Scan(&state.APIKeyID, &state.BudgetUSD, &state.SpentUSD)
	if err != nil {
		return nil, fmt.Errorf("upsert budget: %w", err)
	}

	return &state, nil
}

// DeleteBudget removes a key budget row.
func DeleteBudget(ctx context.Context, db *sql.DB, keyID string) error {
	result, err := db.ExecContext(ctx, `DELETE FROM api_key_budgets WHERE api_key_id = $1`, keyID)
	if err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete budget rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ResetBudget zeroes spent_usd for an existing budget row.
func ResetBudget(ctx context.Context, db *sql.DB, keyID string) (*BudgetState, error) {
	var state BudgetState
	err := db.QueryRowContext(
		ctx,
		`UPDATE api_key_budgets
		 SET spent_usd = 0, updated_at = NOW()
		 WHERE api_key_id = $1
		 RETURNING api_key_id, budget_usd, spent_usd`,
		keyID,
	).Scan(&state.APIKeyID, &state.BudgetUSD, &state.SpentUSD)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reset budget: %w", err)
	}

	return &state, nil
}

// Deduct atomically subtracts costUSD from the key's budget inside a
// serializable transaction. Returns an error if the deduction would
// exceed the budget.
func (bt *PgBudgetTracker) Deduct(ctx context.Context, apiKeyID string, costUSD float64) error {
	tx, err := bt.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var budget, spent float64
	err = tx.QueryRowContext(ctx,
		`SELECT budget_usd, spent_usd FROM api_key_budgets WHERE api_key_id = $1 FOR UPDATE`,
		apiKeyID,
	).Scan(&budget, &spent)

	if err == sql.ErrNoRows {
		// No budget row means unlimited — nothing to deduct from.
		return nil
	}
	if err != nil {
		return fmt.Errorf("select budget for update: %w", err)
	}

	newSpent := spent + costUSD
	if newSpent > budget {
		return fmt.Errorf("budget exhausted for key %s: would be %.4f / %.4f USD", apiKeyID, newSpent, budget)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE api_key_budgets SET spent_usd = $1, updated_at = NOW() WHERE api_key_id = $2`,
		newSpent, apiKeyID,
	)
	if err != nil {
		return fmt.Errorf("update spent: %w", err)
	}

	return tx.Commit()
}
