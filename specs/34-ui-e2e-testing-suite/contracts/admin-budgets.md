# Contract: Admin Budgets

Per-key budget endpoints are documented in [`admin-keys.md`](./admin-keys.md). This file covers team/global budgets surfaced by the budgets page.

## `GET /v1/admin/budgets`

**Response 200**:
```json
{
  "data": [
    {
      "scope": "team",
      "scope_id": "team_research",
      "limit_usd": "1000.00",
      "usage_usd": "860.00",
      "duration": "monthly",
      "alert_threshold": 0.85
    }
  ]
}
```

`alert_threshold` drives the warning indicator (`data-testid="team-{id}-warn"`).

## `PATCH /v1/admin/budgets/{scope}/{scope_id}`

Updates `limit_usd`, `duration`, `alert_threshold`.
