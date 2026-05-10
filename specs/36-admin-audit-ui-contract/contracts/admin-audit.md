# Contract: Admin Audit (production)

## Endpoint

- `GET /v1/admin/audit`

## Authentication

Same admin-route protection as the existing endpoint. No new authentication primitives are introduced by this feature.

## Query Parameters

All parameters are optional. Parameters with deprecated aliases honor the new name when both are present and reject ambiguity with `invalid_request_error`.

| Name          | Type     | Description                                                                                                          | Notes |
|---------------|----------|----------------------------------------------------------------------------------------------------------------------|-------|
| `actor`       | string   | Actor identity. Matches `audit_log.api_key_id` for request-action rows.                                              | Preferred over `api_key_id`. |
| `api_key_id`  | string   | Deprecated alias of `actor`.                                                                                         | Returns 400 if combined with `actor`. |
| `action`      | string   | `request:` family prefix or `request:{model}` exact selector.                                                        | Coexists with `model` (AND). |
| `model`       | string   | Model identifier.                                                                                                    | Existing column filter. |
| `provider`    | string   | Provider display label.                                                                                              | Existing column filter. |
| `outcome`     | string   | One of `success`, `client_error`, `unauthorized`, `rate_limited`, `budget_denied`, `failure`.                        | Preferred over `status`. |
| `status`      | string   | Deprecated. `success` or `error`. Wins-loses to `outcome` when both are set.                                          | |
| `from`        | RFC3339  | Inclusive lower bound on `created_at`.                                                                               | Must be ≤ `to`. |
| `to`          | RFC3339  | Inclusive upper bound on `created_at`.                                                                               | |
| `limit`       | integer  | 1..1000. Default 100.                                                                                                | |
| `offset`      | integer  | ≥ 0. Default 0.                                                                                                      | |

## Success Response

`HTTP 200 OK`

```json
{
  "data": [
    {
      "id": 1024,
      "request_id": "req_abc123",
      "actor_id": "key_checkout_service",
      "api_key_id": "key_checkout_service",
      "action": "request:gpt-4o",
      "model": "gpt-4o",
      "provider": "OpenAI · prod",
      "prompt_tokens": 120,
      "output_tokens": 80,
      "total_tokens": 200,
      "cost_usd": 0.0024,
      "status_code": 200,
      "latency_ms": 432,
      "streaming": false,
      "outcome": "success",
      "error_message": "",
      "created_at": "2026-05-09T11:59:50Z"
    }
  ],
  "total": 1,
  "limit": 100,
  "offset": 0,
  "page": 1,
  "has_more": false
}
```

`reference_summary` is included on rows where one or more references could not be resolved cheaply:

```json
{
  "id": 2048,
  "request_id": "req_def456",
  "actor_id": "",
  "api_key_id": "",
  "action": "request:legacy-model-v1",
  "model": "legacy-model-v1",
  "provider": "Anthropic · prod",
  "outcome": "client_error",
  "error_message": "model not found",
  "reference_summary": [
    { "field": "actor_id", "original_id": "", "state": "missing" },
    { "field": "model",    "original_id": "legacy-model-v1", "state": "unknown" }
  ]
}
```

## Error Response

All validation and runtime failures use the OpenAI-compatible error envelope:

```json
{
  "error": {
    "message": "outcome must be one of: success, client_error, unauthorized, rate_limited, budget_denied, failure",
    "type": "invalid_request_error",
    "code": "invalid_outcome"
  }
}
```

Status codes:

- `400 invalid_request_error` — malformed `from`/`to`, unknown `outcome`, both `actor` and `api_key_id` set, `from > to`, non-positive `limit`, negative `offset`.
- `503 service_unavailable` — database is unreachable.
- `500 internal_error` — unexpected failure. The response message is generic; the underlying database error MUST NOT leak.

## Redaction Rules

`error_message` is redacted at response build time. The redactor replaces detected substrings with `[REDACTED]`:

- HTTP `Authorization` header values, including `Bearer ...` tokens.
- API key prefixes such as `sk-...` or `pk-...`.
- Cookie values.
- Long alphanumeric/base64 substrings (≥ 20 characters).
- Case-insensitive standalone words: `key`, `secret`, `token`, `password`, `credential`.

Surrounding human-readable text MUST be preserved so that:

- The `budget_denied` outcome rule (which matches the keyword `budget` in the message) continues to work after redaction.
- Operators retain enough context to investigate.

## Pagination Rules

- `total` is the exact count of rows matching the active filter at query time.
- Sort order is `created_at DESC, id DESC` (deterministic across millisecond ties).
- For any filter, requesting `offset = k * limit` for `k = 0, 1, 2, ...` MUST yield disjoint, contiguous slices that cover the entire result set without gaps or duplicates.
- `page` and `has_more` are derived fields (no extra round-trips).

## Compatibility

- Existing clients that use only `api_key_id`, `model`, `provider`, `status`, `from`, `to`, `limit`, and `offset` continue to receive functionally identical results. Their response payloads gain new fields (`actor_id`, `action`, `outcome`, `page`, `has_more`, optionally `reference_summary`) but no fields are removed.
- The response payload is additive over feature 35's UI contract; the UI does not need to change to consume the production version. Any unknown fields the UI ignores remain ignored.
