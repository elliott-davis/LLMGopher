# Spec 05: API Key Management Improvements

## Status
completed

## Goal
Extend API key management with four capabilities that are essential for production operation: hard delete, key expiration, metadata/tags for cost attribution, and per-key model allowlists.

## Background
`internal/api/admin.go` handles key CRUD. The `api_keys` table has `id`, `key_hash`, `name`, `rate_limit_rps`, `is_active`, `created_at`, `updated_at`. The `llm.APIKeyConfig` type mirrors this schema.

Current gaps:
- No DELETE route for keys (only soft-deactivation via `is_active`)
- No expiration — keys are valid indefinitely
- No metadata — no way to tag keys for cost attribution or tenant tracking
- No model restriction — any key can call any model

Auth middleware reads the state cache (`internal/middleware/auth.go`) using `AuthWithStateCache`.

## Requirements

### 1. Migration (`internal/storage/migrations/00004_api_key_enhancements.sql`)
```sql
ALTER TABLE api_keys
  ADD COLUMN expires_at TIMESTAMPTZ,
  ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}',
  ADD COLUMN allowed_models TEXT[];
```

### 2. Update `llm.APIKeyConfig` (`pkg/llm/types.go`)
```go
ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
Metadata      map[string]string `json:"metadata,omitempty"`
AllowedModels []string          `json:"allowed_models,omitempty"`
```

### 3. Key expiration enforcement (`internal/middleware/auth.go`)
In `AuthWithStateCache`, after finding the key in the cache, check:
```go
if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
    // return 401 authentication_error "API key has expired"
}
```

### 4. Key expiration in state cache (`internal/storage/cache.go`)
The cache poll query already loads all columns — add `expires_at`, `metadata`, `allowed_models` to the SELECT and scan.

### 5. Model allowlist enforcement (`internal/proxy/handler.go`)
After resolving the provider (before dispatching), check:
```go
if len(key.AllowedModels) > 0 && !slices.Contains(key.AllowedModels, req.Model) {
    // return 403 "model not allowed for this API key"
}
```
Get the key from context via `middleware.GetAPIKeyID` and look it up in the state cache.

### 6. Admin API changes (`internal/api/admin.go`)

**Create key request** — add optional fields:
```go
type createAPIKeyRequest struct {
    Name          string            `json:"name"`
    RateLimitRPS  int               `json:"rate_limit_rps"`
    ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
    AllowedModels []string          `json:"allowed_models,omitempty"`
}
```

**DELETE /v1/admin/keys/{id}** — new route:
- Hard-delete the row from `api_keys` by ID
- Also delete related row from `api_key_budgets`
- Return 204 No Content on success
- Return 404 if key not found

**PUT /v1/admin/keys/{id}** — new route for updating mutable fields:
- Accepts: `name`, `rate_limit_rps`, `expires_at`, `metadata`, `allowed_models`, `is_active`
- Updates the DB row and returns the updated key (without the raw key value — that is never shown again)

### 7. State cache refresh
After a key is deleted or updated via the admin API, the state cache will pick up the change within its polling interval (5 seconds). No immediate invalidation is required.

## Out of Scope
- Key rotation (generating a new secret while preserving the same key ID)
- Bulk key operations
- Allowed endpoints per key (model-level is sufficient for now)

## Acceptance Criteria
- [x] `DELETE /v1/admin/keys/{id}` removes the key from DB and it stops working within 5 seconds
- [x] `PUT /v1/admin/keys/{id}` updates the key's fields
- [x] A key with `expires_at` in the past returns 401
- [x] A key with `allowed_models: ["gpt-4o"]` returns 403 when calling `claude-3-5-sonnet`
- [x] `metadata` is persisted and returned in key list/create responses
- [x] Migration runs cleanly on an existing schema
- [x] Existing keys (no `expires_at`, no `allowed_models`) continue to work unchanged

## Key Files
- `internal/storage/migrations/00004_api_key_enhancements.sql` — new migration (using next available migration number)
- `pkg/llm/types.go` — extend `APIKeyConfig`
- `internal/storage/cache.go` — update cache poll query and scan
- `internal/middleware/auth.go` — expiry check
- `internal/proxy/handler.go` — model allowlist check
- `internal/proxy/handler_completions.go` — model allowlist check for completions
- `internal/proxy/handler_embeddings.go` — model allowlist check for embeddings
- `internal/api/admin.go` — new delete/update routes and updated create
- `internal/api/router.go` — wire new routes
