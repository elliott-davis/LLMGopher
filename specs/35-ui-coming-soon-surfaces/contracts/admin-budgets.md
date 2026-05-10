# Contract: Admin Budgets Surface

## Endpoints

- `GET /v1/admin/budgets`
- `PATCH /v1/admin/budgets/{scope}/{scope_id}` only after production mutation contract reconciliation or in mock-backed E2E

The first implementation may use the existing mock contract from `specs/34-ui-e2e-testing-suite/contracts/admin-budgets.md` and key budget fixtures from `ui/tests/fixtures/keys.ts`.

## Response

```json
{
  "data": [
    {
      "scope": "team",
      "scope_id": "team_research",
      "display_name": "Research",
      "limit_usd": 1000.0,
      "usage_usd": 860.0,
      "duration": "monthly",
      "alert_threshold": 0.85,
      "hard_cap_state": "near_cap"
    }
  ]
}
```

## Editable Request

```json
{
  "limit_usd": 1200.0,
  "duration": "monthly",
  "alert_threshold": 0.85
}
```

## UI Rules

- `limit_usd` and `usage_usd` must be non-negative.
- `alert_threshold` must be greater than 0 and less than or equal to 1.
- Policy edits must preserve current usage values.
- Save failure must preserve entered values and display the gateway error reason.
