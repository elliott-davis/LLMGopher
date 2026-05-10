# Contract: Admin Providers

Mirrors the real `/v1/admin/providers` gateway endpoints.

## `GET /v1/admin/providers`

**Response 200**:
```json
{
  "data": [
    {
      "id": "prov_openai_prod",
      "kind": "openai",
      "display_name": "OpenAI · prod",
      "base_url": "https://api.openai.com",
      "health": "healthy",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

## `POST /v1/admin/providers`

**Request**:
```json
{
  "kind": "openai",
  "display_name": "OpenAI · prod",
  "api_key": "sk-test-...",
  "base_url": "https://api.openai.com"
}
```

**Validation**:
- `kind` MUST be one of the supported provider kinds.
- `display_name` MUST be non-empty.
- `base_url` MUST parse as a URL when provided; invalid URL → `400 invalid_request_error`.

**Response 201**: the created provider object.

## `PATCH /v1/admin/providers/{id}` and `DELETE /v1/admin/providers/{id}`

Standard semantics. `DELETE` cascades to models attached to the provider — return `409 invalid_request_error` if any active key references one of those models.
