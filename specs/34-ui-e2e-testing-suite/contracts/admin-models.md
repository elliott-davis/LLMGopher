# Contract: Admin Models

Mirrors `/v1/admin/models`.

## `GET /v1/admin/models`

**Response 200**:
```json
{
  "data": [
    {
      "id": "gpt-4o",
      "provider_id": "prov_openai_prod",
      "enabled": true,
      "rate_limit": { "rps": 100, "tpm": 60000 } 
    }
  ]
}
```

`rate_limit` is optional.

## `POST /v1/admin/models`, `PATCH /v1/admin/models/{id}`, `DELETE /v1/admin/models/{id}`

Standard CRUD. `PATCH` is partial.
