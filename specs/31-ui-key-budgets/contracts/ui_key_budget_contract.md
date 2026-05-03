# Contract: UI API Key Budget Controls

## Scope

The admin UI exposes existing key-scoped budget capabilities from the gateway. The gateway remains the source of truth for budget persistence, spend counters, reset behavior, and validation.

## Authentication

Budget endpoints require a Bearer token accepted by the gateway auth middleware. UI code must attach this header only from server-side code:

```text
Authorization: Bearer <server-only-admin-api-key>
```

The token must not be passed to client components, stored in browser state, logged, or rendered. If the token is missing or rejected, the UI shows budget controls as unavailable with actionable configuration feedback.

## Gateway Endpoints

### Get Budget

```http
GET /v1/admin/keys/{id}/budget
```

**Success response**: `200 OK`

```json
{
  "api_key_id": "key-001",
  "budget_usd": 100,
  "spent_usd": 25.5,
  "remaining_usd": 74.5,
  "alert_threshold_pct": 80,
  "budget_duration": "monthly",
  "budget_reset_at": "2026-06-01T00:00:00Z"
}
```

**No budget response**: `404 Not Found`

The UI treats this as an unbudgeted key state, not as a blocking page error.

### Create Or Update Budget

```http
PUT /v1/admin/keys/{id}/budget
Content-Type: application/json
```

**Request body**:

```json
{
  "budget_usd": 100,
  "alert_threshold_pct": 80,
  "budget_duration": "monthly",
  "budget_reset_at": "2026-06-01T00:00:00Z"
}
```

**Rules**:

- `budget_usd` is required and must be greater than zero.
- `alert_threshold_pct` is optional and must be 1 through 99 when provided.
- `budget_duration` is optional and must be `daily`, `weekly`, or `monthly` when provided.
- `budget_reset_at` is required when `budget_duration` is provided.
- Existing `spent_usd` is preserved by update unless reset is explicitly invoked.

**Success response**: `200 OK` with the budget response shape.

### Reset Budget Spend

```http
POST /v1/admin/keys/{id}/budget/reset
```

**Success response**: `200 OK` with the budget response shape and `spent_usd` reset to zero.

The UI must require explicit confirmation before calling this endpoint.

### Remove Budget

```http
DELETE /v1/admin/keys/{id}/budget
```

**Success response**: `204 No Content`

The UI must require explicit confirmation before calling this endpoint and then show the unbudgeted state.

## UI Server Actions

Planned server-side actions:

- `fetchAPIKeyBudget(apiKeyID)` returns budget state, unbudgeted state, or unavailable state.
- `upsertAPIKeyBudget(apiKeyID, formData)` validates form data, calls `PUT`, and revalidates `/keys`.
- `resetAPIKeyBudget(apiKeyID)` calls `POST .../reset` and revalidates `/keys`.
- `deleteAPIKeyBudget(apiKeyID)` calls `DELETE` and revalidates `/keys`.

All actions must parse gateway error envelopes with the existing action helper pattern and return or throw administrator-readable messages.

## UI Behavior

- `/keys` remains the primary entry point.
- Each key row shows budget status or a manage-budget action.
- The budget management UI displays limit, spent, remaining, threshold, duration, and reset time when available.
- Invalid form submissions are blocked before save with field-specific feedback.
- Successful mutations refresh budget state after the save/reset/remove call.
- Backend or auth failures do not hide the API key row; they show budget state as unavailable.

## Compatibility

- Existing key creation, update, activation, and deletion flows remain unchanged.
- Existing gateway budget API behavior remains unchanged.
- No API-only budget lifecycle capability remains after this feature, except broader budget roadmap items outside the key-scoped spec 07 API.
