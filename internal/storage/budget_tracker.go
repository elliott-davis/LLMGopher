package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/ed007183/llmgopher/pkg/llm"
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
type BudgetState = llm.BudgetState

// GetBudget returns the current budget state for a key.
// A missing row returns (nil, nil).
func (bt *PgBudgetTracker) GetBudget(ctx context.Context, apiKeyID string) (*BudgetState, error) {
	return GetBudget(ctx, bt.db, apiKeyID)
}

// MarkBudgetAlerted records when a threshold alert was emitted for an API key.
func (bt *PgBudgetTracker) MarkBudgetAlerted(ctx context.Context, apiKeyID string, alertedAt time.Time) error {
	return MarkBudgetAlerted(ctx, bt.db, apiKeyID, alertedAt)
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
		`SELECT
			api_key_id,
			budget_usd,
			spent_usd,
			COALESCE(alert_threshold_pct, 0),
			COALESCE(budget_duration, ''),
			budget_reset_at,
			last_alerted_at
		 FROM api_key_budgets
		 WHERE api_key_id = $1`,
		keyID,
	).Scan(
		&state.APIKeyID,
		&state.BudgetUSD,
		&state.SpentUSD,
		&state.AlertThresholdPct,
		&state.BudgetDuration,
		&state.BudgetResetAt,
		&state.LastAlertedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get budget: %w", err)
	}

	return &state, nil
}

// UpsertBudget creates or updates a budget row while preserving existing spent_usd.
func UpsertBudget(
	ctx context.Context,
	db *sql.DB,
	keyID string,
	budgetUSD float64,
	alertThresholdPct *int,
	budgetDuration *string,
	budgetResetAt *time.Time,
) (*BudgetState, error) {
	var state BudgetState
	err := db.QueryRowContext(
		ctx,
		`INSERT INTO api_key_budgets (
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
		    last_alerted_at`,
		keyID,
		budgetUSD,
		alertThresholdPct,
		budgetDuration,
		budgetResetAt,
	).Scan(
		&state.APIKeyID,
		&state.BudgetUSD,
		&state.SpentUSD,
		&state.AlertThresholdPct,
		&state.BudgetDuration,
		&state.BudgetResetAt,
		&state.LastAlertedAt,
	)
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
		 SET spent_usd = 0, last_alerted_at = NULL, updated_at = NOW()
		 WHERE api_key_id = $1
		 RETURNING
		    api_key_id,
		    budget_usd,
		    spent_usd,
		    COALESCE(alert_threshold_pct, 0),
		    COALESCE(budget_duration, ''),
		    budget_reset_at,
		    last_alerted_at`,
		keyID,
	).Scan(
		&state.APIKeyID,
		&state.BudgetUSD,
		&state.SpentUSD,
		&state.AlertThresholdPct,
		&state.BudgetDuration,
		&state.BudgetResetAt,
		&state.LastAlertedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reset budget: %w", err)
	}

	return &state, nil
}

// MarkBudgetAlerted sets last_alerted_at for a budget row.
func MarkBudgetAlerted(ctx context.Context, db *sql.DB, keyID string, alertedAt time.Time) error {
	result, err := db.ExecContext(
		ctx,
		`UPDATE api_key_budgets SET last_alerted_at = $1, updated_at = NOW() WHERE api_key_id = $2`,
		alertedAt,
		keyID,
	)
	if err != nil {
		return fmt.Errorf("mark budget alerted: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark budget alerted rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// StartBudgetResetWorker polls every minute and resets due recurring budgets.
func StartBudgetResetWorker(ctx context.Context, db *sql.DB, logger *slog.Logger) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resetIDs, err := resetDueBudgets(ctx, db)
			if err != nil {
				logger.Error("budget reset worker failed", "error", err)
				continue
			}
			for _, id := range resetIDs {
				logger.Info("budget_reset_completed", "api_key_id", id)
			}
		}
	}
}

func resetDueBudgets(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(
		ctx,
		`UPDATE api_key_budgets
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
		 RETURNING api_key_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("reset due budgets: %w", err)
	}
	defer rows.Close()

	var keyIDs []string
	for rows.Next() {
		var keyID string
		if err := rows.Scan(&keyID); err != nil {
			return nil, fmt.Errorf("scan reset budget row: %w", err)
		}
		keyIDs = append(keyIDs, keyID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reset budget rows: %w", err)
	}

	return keyIDs, nil
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
