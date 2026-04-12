# Spec 03: GET /v1/models Endpoint

## Status
complete

## Goal
Expose an OpenAI-compatible `GET /v1/models` endpoint that returns the list of configured models. Many OpenAI SDK clients call this endpoint on initialization to discover available models; without it, those clients fail before making any requests.

## Background
The state cache (`internal/storage/cache.go`) already holds the full model list, keyed by alias. The admin endpoint `GET /v1/admin/models` returns this same data in an internal format. The new endpoint returns the OpenAI-specified format.

The OpenAI models list response format:
```json
{
  "object": "list",
  "data": [
    {"id": "gpt-4o", "object": "model", "created": 1234567890, "owned_by": "llmgopher"},
    ...
  ]
}
```

## Requirements

1. Add route `GET /v1/models` in `internal/api/router.go`, behind the Auth middleware (same as other `/v1/` endpoints).

2. The handler reads the state cache and returns all active models. For each `ModelConfig`:
   - `id` = `Alias` (the name clients use when calling the API)
   - `object` = `"model"`
   - `created` = Unix timestamp from `CreatedAt`
   - `owned_by` = `"llmgopher"`

3. If the state cache is nil or empty, return an empty `data` array (not an error).

4. No authentication beyond what the existing Auth middleware already enforces.

## Out of Scope
- `GET /v1/models/{model}` (single model detail) — add later if needed
- Dynamic provider model discovery (listing all models from upstream providers)
- Filtering by provider or capability

## Acceptance Criteria
- [x] `GET /v1/models` returns HTTP 200 with `{"object": "list", "data": [...]}`
- [x] Each entry has `id`, `object`, `created`, `owned_by`
- [x] The `id` field matches the alias used when calling `/v1/chat/completions`
- [x] Returns empty data array when no models are configured
- [x] Requires a valid API key (auth middleware enforced)
- [x] OpenAI Python SDK `client.models.list()` call succeeds against the gateway _(response shape verified in tests; optional manual smoke test with a real client)_

## Key Files
- `internal/api/router.go` — add route
- `internal/api/admin.go` — add handler (or create `internal/api/models.go`)
