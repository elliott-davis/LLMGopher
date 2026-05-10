# Contract: Admin Teams Surface

## Endpoint

- `GET /v1/admin/teams`

The first implementation may use the existing mock contract from `specs/34-ui-e2e-testing-suite/contracts/admin-teams.md`. Production fields must be reconciled with `23-teams-organizations` before enabling production data beyond the read-only overview.

## Response

```json
{
  "data": [
    {
      "id": "team_research",
      "display_name": "Research",
      "member_count": 8,
      "budget_utilization": 0.86,
      "budget_health": "near_cap"
    }
  ]
}
```

## UI Rules

- `member_count` and `budget_utilization` must be non-negative.
- Teams at or above their configured alert threshold must be marked with text and visual treatment.
- Empty teams and unavailable teams service states must be distinct.
- Team creation, deletion, renaming, membership, and RBAC role assignment are out of scope.
