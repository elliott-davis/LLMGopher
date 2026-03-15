package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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
