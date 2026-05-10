# Contract: Admin Rate Limits Surface

## Endpoints

- `GET /v1/admin/rate-limits`
- `POST /v1/admin/rate-limits` only after production mutation contract reconciliation or in mock-backed E2E
- `PATCH /v1/admin/rate-limits/{id}` only after production mutation contract reconciliation or in mock-backed E2E
- `DELETE /v1/admin/rate-limits/{id}` only after production mutation contract reconciliation or in mock-backed E2E

The first implementation may use the existing mock contract from `specs/34-ui-e2e-testing-suite/contracts/admin-rate-limits.md`.

## Response

```json
{
  "data": [
    {
      "id": "rl_tripped",
      "scope": "team",
      "scope_id": "team_research",
      "rps": 5,
      "tpm": 20000,
      "tripped": true
    }
  ]
}
```

## Editable Request

```json
{
  "scope": "model",
  "scope_id": "gpt-4o-mini",
  "rps": 10,
  "tpm": 30000
}
```

## UI Rules

- `scope` values are `model`, `key`, or `team`.
- `rps` and `tpm`, when present, must be non-negative.
- At least one meaningful limit is required.
- Token-per-minute controls must be marked unavailable if production enforcement is not confirmed.
- Tripped rules must be discoverable through accessible text, not color alone.
