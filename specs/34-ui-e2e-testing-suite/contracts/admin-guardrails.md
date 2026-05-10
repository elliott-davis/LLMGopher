# Contract: Admin Guardrails

Real guardrails API is not yet shipped (see spec `26-guardrail-integrations`). Mock contract:

## `GET /v1/admin/guardrails`

**Response 200**:
```json
{
  "data": [
    { "id": "gr_jail",    "display_name": "Jailbreak detection", "enabled": false },
    { "id": "gr_pii",     "display_name": "PII redaction",        "enabled": true  },
    { "id": "gr_secrets", "display_name": "Secret scanning",      "enabled": false }
  ]
}
```

## `PATCH /v1/admin/guardrails/{id}`

**Request**: `{ "enabled": true }`. Persistence MUST survive across reloads in the same Playwright worker.
