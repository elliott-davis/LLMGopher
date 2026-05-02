# Contract: UI API Key Lifecycle Controls

## Scope

This contract documents how the admin UI uses existing gateway admin APIs to manage API key lifecycle state. It does not introduce new backend endpoints.

## Data Sources

### List Keys

`GET /v1/admin/keys`

Returns the live key cache snapshot used by the key inventory.

Expected UI fields:

- `id`
- `key_hash`
- `name`
- `rate_limit_rps`
- `is_active`
- `expires_at`
- `metadata`
- `allowed_models`
- `created_at`
- `updated_at`

### List Models

`GET /v1/admin/models`

Returns model choices used to render model allowlists with human-readable labels while preserving exact identifiers.

Expected UI fields:

- `id`
- `name`
- `alias`
- `provider_id`
- `context_window`
- `created_at`
- `updated_at`

## Mutations

### Create Key

`POST /v1/admin/keys`

Request body:

```json
{
  "name": "Production Service Key",
  "rate_limit_rps": 100,
  "expires_at": "2026-12-31T23:59:59Z",
  "metadata": {
    "owner": "platform"
  },
  "allowed_models": ["gpt-4o"]
}
```

Response body includes the raw key exactly once:

```json
{
  "id": "11111111-1111-1111-1111-111111111111",
  "name": "Production Service Key",
  "rate_limit_rps": 100,
  "is_active": true,
  "expires_at": "2026-12-31T23:59:59Z",
  "metadata": {
    "owner": "platform"
  },
  "allowed_models": ["gpt-4o"],
  "api_key": "sk-..."
}
```

UI obligations:

- Display `api_key` only in the successful create dialog.
- Clear raw key display when the dialog closes.
- Refresh `/keys` after success.

### Update Key

`PUT /v1/admin/keys/{id}`

Request body may include any mutable fields:

```json
{
  "name": "Renamed Key",
  "rate_limit_rps": 50,
  "expires_at": null,
  "metadata": {
    "owner": "security"
  },
  "allowed_models": [],
  "is_active": false
}
```

Response body returns the updated API key without raw key material.

UI obligations:

- Preserve form state on validation or gateway failure.
- Treat `is_active: false` as deactivate and `is_active: true` as reactivate.
- Show empty `allowed_models` as unrestricted.
- Refresh `/keys` after success.

### Delete Key

`DELETE /v1/admin/keys/{id}`

Success response: `204 No Content`

UI obligations:

- Require explicit confirmation before submission.
- Explain that deletion is permanent and clients using the key will stop authenticating after synchronization.
- Remove the key from the inventory after refresh or show a waiting state if synchronization takes longer than expected.

## Error Handling

Gateway errors use the existing envelope:

```json
{
  "error": {
    "message": "api key not found",
    "type": "invalid_request_error",
    "code": "invalid_request_error"
  }
}
```

UI obligations:

- Prefer the gateway `error.message` when available.
- Prefix or contextualize messages by action, such as "Failed to update API key".
- Never include raw key material in errors, logs, toast messages, or table cells.

## API-Only Exceptions

No key lifecycle capability from this feature remains intentionally API-only. Key rotation that preserves the same key ID remains out of scope because the backend does not persist plaintext keys and the feature spec excludes it.

## Implemented UI Behavior Notes

- The key inventory fetches keys and models with `cache: "no-store"` so lifecycle fields and model labels are server-authoritative.
- Create and edit forms submit `expires_at`, `metadata`, and `allowed_models`; empty allowlists are rendered and submitted as unrestricted access.
- Stale allowed model identifiers are preserved in the edit UI and table labels rather than silently removed.
- Deactivate/reactivate use `PUT /v1/admin/keys/{id}` with `is_active` only, preserving the same key ID.
- Delete uses a confirmation dialog before `DELETE /v1/admin/keys/{id}` and waits for refreshed inventory state before reporting sync completion.
