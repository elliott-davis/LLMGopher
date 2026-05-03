# Data Model: UI API Key Budget Controls

## APIKeyBudget

Represents the budget state returned by the gateway for one API key.

**Fields**:

- `api_key_id`: string UUID identifying the key.
- `budget_usd`: positive decimal budget limit.
- `spent_usd`: non-negative decimal spend counted against the budget.
- `remaining_usd`: decimal remaining balance, derived by gateway state.
- `alert_threshold_pct`: optional integer from 1 to 99.
- `budget_duration`: optional enum: `daily`, `weekly`, `monthly`.
- `budget_reset_at`: optional ISO timestamp; required when `budget_duration` is present.

**Validation rules**:

- `budget_usd` must be greater than zero.
- `spent_usd` and `remaining_usd` are display-only in the UI.
- `alert_threshold_pct` must be empty or between 1 and 99.
- `budget_duration` must be empty, `daily`, `weekly`, or `monthly`.
- `budget_reset_at` is required when `budget_duration` is set.

## APIKeyBudgetForm

Represents administrator input for creating or updating a budget.

**Fields**:

- `budget_usd`: string input parsed to a positive number.
- `alert_threshold_pct`: optional string input parsed to an integer.
- `budget_duration`: optional duration enum.
- `budget_reset_at`: optional datetime-local value converted to an ISO timestamp.

**Validation rules**:

- Reject missing, non-numeric, zero, or negative budget limit.
- Reject threshold values outside 1 through 99.
- Reject unsupported duration values.
- Reject duration-based submissions without reset time.
- Preserve current spend when saving unless the administrator invokes the reset action.

## BudgetStatusIndicator

Represents the UI state shown on the key management surface.

**States**:

- `unbudgeted`: no budget row exists for the key.
- `within_budget`: a budget exists and spend is below alert threshold or no threshold is configured.
- `near_threshold`: spend has reached or exceeded the configured threshold percentage and remaining budget is positive.
- `exhausted`: remaining budget is zero or negative.
- `unavailable`: budget state cannot be fetched due to backend, network, auth, or configuration errors.

**Validation rules**:

- Treat gateway `404` for budget lookup as `unbudgeted`.
- Treat `401`/`403` or missing server token configuration as `unavailable` with an admin setup message.
- Calculate display percentages from returned `budget_usd` and `spent_usd`, guarding against invalid response values.

## BudgetLifecycleAction

Represents a mutation initiated by an administrator.

**Types**:

- `set`: create budget for an unbudgeted key.
- `update`: modify budget limit, threshold, duration, or reset timing.
- `reset`: reset spend counters to zero.
- `remove`: delete the key budget.

**Validation rules**:

- `set` and `update` require a valid `APIKeyBudgetForm`.
- `reset` requires explicit confirmation and must refetch budget state after success.
- `remove` requires explicit confirmation and must return the UI to the unbudgeted state after success.
- All mutations must surface gateway errors through the existing OpenAI-compatible error envelope parser where possible.

## Relationships

- One API key has zero or one `APIKeyBudget`.
- One `APIKeyBudgetForm` creates or updates one `APIKeyBudget`.
- One `BudgetStatusIndicator` summarizes one API key's budget state.
- A `BudgetLifecycleAction` targets one API key and may produce a new `APIKeyBudget` state or the unbudgeted state.

## State Transitions

```text
unbudgeted --set--> within_budget | near_threshold | exhausted
within_budget --update--> within_budget | near_threshold | exhausted
near_threshold --update--> within_budget | near_threshold | exhausted
exhausted --update--> within_budget | near_threshold | exhausted
within_budget|near_threshold|exhausted --reset--> within_budget
within_budget|near_threshold|exhausted --remove--> unbudgeted
any state --fetch/auth/backend failure--> unavailable
unavailable --successful refetch--> unbudgeted | within_budget | near_threshold | exhausted
```
