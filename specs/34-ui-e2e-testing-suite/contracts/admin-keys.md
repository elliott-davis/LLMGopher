# Contract: Admin API Keys

Mirrors `/v1/admin/keys` and the budget sub-resource.

## `GET /v1/admin/keys`

**Response 200**: `{ "data": [APIKey, ...] }` where `APIKey` matches `ui/src/lib/types.ts`.

The mock MUST NOT return any `key` field with the full secret; only `prefix` (last 4 chars).

## `POST /v1/admin/keys`

**Response 201**: the created key including a one-time `key` field with the full secret. The reveal is shown exactly once in the UI; subsequent `GET` calls MUST omit the field.

## `PATCH /v1/admin/keys/{id}`

Editable: `name`, `is_active`, `model_allowlist`, `rate_limit`, `expires_at`, `metadata`.

## `POST /v1/admin/keys/{id}/rotate`

**Response 200**: `{ "id": "...", "key": "<new-secret>", "prefix": "...wxyz" }`.

This endpoint is the target of the `rotate emits exactly one POST` assertion.

## `PUT /v1/admin/keys/{id}/budget`, `POST /v1/admin/keys/{id}/budget/reset`, `DELETE /v1/admin/keys/{id}/budget`

Mirror the existing real-gateway shapes — see `ui/src/lib/budget.ts` for the form-values parser the real UI uses today.

## Cap enforcement (contract test)

`POST /v1/chat/completions` — when the requesting key has `usage_usd >= limit_usd`, MUST return:

```text
HTTP/1.1 429 Too Many Requests
x-llmgopher-reason: budget_exceeded

{"error":{"message":"budget exceeded","type":"rate_limit_error","code":"budget_exceeded"}}
```

This is the only chat-completions behavior the mock implements; everything else 501s.
