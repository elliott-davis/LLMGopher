# Spec 10: Budget Lifecycle (Alerts & Periodic Resets)

## Status
completed

## Goal
Add two missing budget features: (1) soft alerts that warn when a key approaches its limit, and (2) automatic periodic budget resets so keys don't have to be manually reset each billing cycle.

## Background
`internal/storage/budget_tracker.go` handles budget deduction via `Deduct()`. The `CostWorker` in `internal/proxy/cost_worker.go` calls `Deduct()` after each request. The `api_key_budgets` table has `budget_usd`, `spent_usd`, `created_at`, `updated_at`.

Spec 07 adds `reset_at` to the budget row and exposes budget CRUD. This spec depends on spec 07's migration having added the `budget_duration` and `reset_at` columns (or this spec can add them).

## Requirements

### 1. Migration (`internal/storage/migrations/00006_budget_lifecycle.sql`)

```sql
ALTER TABLE api_key_budgets
  ADD COLUMN alert_threshold_pct INTEGER,       -- e.g. 80 means alert at 80% spent
  ADD COLUMN budget_duration TEXT,              -- "daily" | "weekly" | "monthly"
  ADD COLUMN budget_reset_at TIMESTAMPTZ,       -- next scheduled reset time
  ADD COLUMN last_alerted_at TIMESTAMPTZ;       -- prevents repeat alerts
```

### 2. Update `BudgetState` type (`internal/storage/budget_tracker.go`)

Add:
```go
AlertThresholdPct int
BudgetDuration    string
BudgetResetAt     *time.Time
```

Include these in `GetBudget` and `UpsertBudget` storage functions (spec 07).

### 3. Soft budget alerts

In `CostWorker.deductAndLog()` (`internal/proxy/cost_worker.go`), after a successful deduction:

```go
pctSpent := (state.SpentUSD / state.BudgetUSD) * 100
if state.AlertThresholdPct > 0 &&
   pctSpent >= float64(state.AlertThresholdPct) &&
   (state.LastAlertedAt == nil || time.Since(*state.LastAlertedAt) > 1*time.Hour) {
    w.logger.Warn("budget threshold reached",
        "api_key_id", keyID,
        "spent_usd", state.SpentUSD,
        "budget_usd", state.BudgetUSD,
        "pct_spent", pctSpent,
    )
    // Update last_alerted_at to suppress repeat alerts for 1 hour
    db.Exec("UPDATE api_key_budgets SET last_alerted_at = NOW() WHERE api_key_id = $1", keyID)
}
```

Alert is a structured log entry (slog.Warn) with a well-known `msg` value (`"budget_threshold_reached"`) so it can be filtered or forwarded by log aggregation systems. No email or Slack in this spec.

### 4. Periodic budget reset background task

Add a `BudgetResetWorker` in `internal/storage/budget_tracker.go`:

```go
// StartBudgetResetWorker polls every minute and resets budgets whose reset_at has passed.
func StartBudgetResetWorker(ctx context.Context, db *sql.DB, logger *slog.Logger)
```

The worker runs:
```sql
UPDATE api_key_budgets
SET
    spent_usd = 0,
    budget_reset_at = CASE
        WHEN budget_duration = 'daily'   THEN budget_reset_at + INTERVAL '1 day'
        WHEN budget_duration = 'weekly'  THEN budget_reset_at + INTERVAL '1 week'
        WHEN budget_duration = 'monthly' THEN budget_reset_at + INTERVAL '1 month'
    END,
    updated_at = NOW()
WHERE budget_reset_at IS NOT NULL
  AND budget_duration IS NOT NULL
  AND budget_reset_at <= NOW()
RETURNING api_key_id
```

Log each reset as a structured info entry. Start the worker in `cmd/gateway/main.go` alongside the other background workers.

### 5. Admin API changes (spec 07 extension)

`PUT /v1/admin/keys/{id}/budget` accepts the new fields:
```json
{
  "budget_usd": 100.00,
  "alert_threshold_pct": 80,
  "budget_duration": "monthly",
  "budget_reset_at": "2025-02-01T00:00:00Z"
}
```

Validate: `alert_threshold_pct` in range 1–99 if provided; `budget_duration` must be `"daily"`, `"weekly"`, or `"monthly"` if provided; if `budget_duration` is set, `budget_reset_at` is required.

## Out of Scope
- Email or webhook alerts (covered later in spec 22)
- Per-model budgets (spec 23)
- Retroactive reset (only resets at scheduled `budget_reset_at`, not backdated)

## Acceptance Criteria
- [x] A key at 80% of its budget with `alert_threshold_pct: 80` logs a `budget_threshold_reached` warning
- [x] The alert does not fire again within 1 hour for the same key
- [x] A key with `budget_duration: "monthly"` and `budget_reset_at` in the past gets `spent_usd` reset to 0
- [x] `budget_reset_at` advances by one month after reset
- [x] A key without `budget_duration` is never auto-reset
- [x] Migration runs cleanly on existing schema
- [x] BudgetResetWorker starts at gateway startup and stops on graceful shutdown

## Key Files
- `internal/storage/migrations/00006_budget_lifecycle.sql` — new migration
- `internal/storage/budget_tracker.go` — reset worker, updated types and queries
- `internal/proxy/cost_worker.go` — alert logic in deductAndLog
- `internal/api/admin.go` — budget lifecycle request validation and response fields
- `cmd/gateway/main.go` — start reset worker
