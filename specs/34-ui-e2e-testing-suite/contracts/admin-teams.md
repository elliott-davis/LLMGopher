# Contract: Admin Teams

Real teams API is not yet shipped (see spec `23-teams-organizations`). The mock implements the minimum needed for the teams grid.

## `GET /v1/admin/teams`

**Response 200**:
```json
{
  "data": [
    {
      "id": "team_research",
      "display_name": "Research",
      "member_count": 12,
      "budget_utilization": 0.86
    }
  ]
}
```

When the real teams API ships, this contract MUST be reconciled with it before the mock is updated.
