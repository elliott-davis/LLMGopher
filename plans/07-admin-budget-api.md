# Spec 07: Budget Management API

## Status
completed

## Goal
Expose CRUD endpoints for managing per-key budgets. Currently `api_key_budgets` is written to by the cost worker but has no read or write API — operators cannot view or set budgets without direct database access.

## Background
The `api_key_budgets` table has: `api_key_id` (UUID PK, FK to `api_keys`), `budget_usd`, `spent_usd`, `created_at`, `updated_at`.

`internal/storage/budget_tracker.go` has `RemainingBudget(ctx, keyID)` and `Deduct(ctx, keyID, amount)`. Neither exposes a full read of the budget row.

## Requirements

### 1. New endpoint: `GET /v1/admin/keys/{id}/budget`

Returns the current budget state for a key:
```json
{
  "api_key_id": "...",
  "budget_usd": 100.00,
  "spent_usd": 23.45,
  "remaining_usd": 76.55,
  "reset_at": null
}
```

Return 404 if no budget row exists for the key. `reset_at` is null until spec 10 adds budget lifecycle.

### 2. New endpoint: `PUT /v1/admin/keys/{id}/budget`

Creates or replaces the budget for a key (upsert):
```json
{"budget_usd": 100.00}
```

- Validates `budget_usd > 0`
- Does not reset `spent_usd` — only updates the limit
- Returns the updated budget state (same shape as GET)

### 3. New endpoint: `DELETE /v1/admin/keys/{id}/budget`

Removes the budget row, effectively giving the key unlimited spend. Returns 204. Returns 404 if no budget exists.

### 4. New endpoint: `POST /v1/admin/keys/{id}/budget/reset`

Resets `spent_usd` to 0 without changing `budget_usd`. Returns the updated budget state. Useful for manual monthly resets before spec 10 automates this.

### 5. Storage layer

Add to `internal/storage/budget_tracker.go`:
```go
type BudgetState struct {
    APIKeyID  string
    BudgetUSD float64
    SpentUSD  float64
    ResetAt   *time.Time
}

func GetBudget(ctx context.Context, db *sql.DB, keyID string) (*BudgetState, error)
func UpsertBudget(ctx context.Context, db *sql.DB, keyID string, budgetUSD float64) (*BudgetState, error)
func DeleteBudget(ctx context.Context, db *sql.DB, keyID string) error
func ResetBudget(ctx context.Context, db *sql.DB, keyID string) (*BudgetState, error)
```

`GetBudget` returns `nil, nil` (not an error) when no row exists.

### 6. Routes

Add all four routes to `internal/api/router.go` under the existing admin middleware chain.

## Out of Scope
- Budget alerts (spec 10)
- Periodic reset scheduling (spec 10)
- Per-model or per-team budgets (spec 23)

## Acceptance Criteria
- [x] `GET /v1/admin/keys/{id}/budget` returns budget state for an existing key
- [x] `GET` returns 404 when no budget is set
- [x] `PUT /v1/admin/keys/{id}/budget` creates a budget on first call
- [x] `PUT` updates `budget_usd` without touching `spent_usd`
- [x] `DELETE /v1/admin/keys/{id}/budget` removes the budget row
- [x] `POST /v1/admin/keys/{id}/budget/reset` sets `spent_usd` to 0
- [x] `remaining_usd` = `budget_usd` - `spent_usd` is computed correctly in all responses
- [x] `budget_usd` <= 0 returns 400

## Key Files
- `internal/storage/budget_tracker.go` — add storage functions
- `internal/api/admin.go` — new handlers
- `internal/api/router.go` — new routes
