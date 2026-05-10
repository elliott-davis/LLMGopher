# Contract: Admin Guardrails Surface

## Endpoints

- `GET /v1/admin/guardrails`
- `PATCH /v1/admin/guardrails/{id}` only after production mutation contract reconciliation or in mock-backed E2E

The first implementation may use the existing mock contract from `specs/34-ui-e2e-testing-suite/contracts/admin-guardrails.md`.

## Response

```json
{
  "data": [
    {
      "id": "gr_jail",
      "display_name": "Jailbreak Detection",
      "enabled": false,
      "category": "prompt",
      "description": "Detects jailbreak attempts",
      "provider_label": "Built-in",
      "last_updated_at": "2026-05-09T00:00:00Z"
    }
  ]
}
```

## Toggle Request

```json
{
  "enabled": true
}
```

## UI Rules

- Toggle state must show saving, success, and failure states in editable contexts.
- Failed saves must not leave final enabled state ambiguous.
- Raw detector prompts, match payloads, provider credentials, and sensitive response content are never rendered.
