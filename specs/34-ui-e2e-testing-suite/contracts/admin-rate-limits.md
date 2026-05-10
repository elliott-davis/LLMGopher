# Contract: Admin Rate Limits

Aligns with spec `33-ui-model-rate-limits` (currently in flight on the working branch). When that feature ships its real endpoints, this mock contract MUST be reconciled.

## `GET /v1/admin/rate-limits`

**Response 200**:
```json
{
  "data": [
    {
      "id": "rl_chat_default",
      "scope": "model",
      "scope_id": "gpt-4o",
      "rps": 100,
      "tpm": 60000,
      "tripped": false
    }
  ]
}
```

`tripped: true` flips a "tripped" pill in the UI. The fixture seed sets exactly one rule tripped.

## `POST /v1/admin/rate-limits`, `PATCH /v1/admin/rate-limits/{id}`, `DELETE /v1/admin/rate-limits/{id}`

Standard CRUD.
