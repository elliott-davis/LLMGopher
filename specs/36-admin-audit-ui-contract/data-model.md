# Phase 1 Data Model: Admin Audit UI Contract

## Audit Record (response payload)

The wire shape returned to admin clients for each audit row.

- `id`: stable row identifier (`int64`).
- `request_id`: gateway request/call identifier (`string`).
- `actor_id`: actor-like identity. For request-action rows, this is the `api_key_id`; for future admin-action rows it MAY be a stable administrator identifier (`string`).
- `api_key_id`: deprecated alias of `actor_id`, retained for backward compatibility (`string`). Existing admin clients SHOULD migrate to `actor_id`; the field MUST continue to be present until at least one minor version after this feature ships.
- `action`: classifier of the audited operation. For request rows, `request:{model}` (e.g., `request:gpt-4o`). Future admin operations MAY use other prefixes such as `admin:keys.create`.
- `model`: requested or resolved model identifier (`string`).
- `provider`: provider display label (`string`).
- `prompt_tokens`, `output_tokens`, `total_tokens`: token counts (`int`).
- `cost_usd`: gateway-calculated cost for this row (`float64`).
- `status_code`: HTTP status of the original request (`int`).
- `latency_ms`: request latency in milliseconds (`int64`).
- `streaming`: whether the original response streamed (`bool`).
- `outcome`: derived label (`string`, one of `success`, `client_error`, `unauthorized`, `rate_limited`, `budget_denied`, `failure`).
- `error_message`: redacted summary of the error path; safe substrings only (`string`).
- `reference_summary`: optional `ReferenceSummary[]` describing missing-or-deleted referenced entities (omitted when all references are intact).
- `created_at`: ISO 8601 timestamp (`time.Time`).

Validation rules:

- `id` and `request_id` MUST be present on every row.
- `outcome` MUST be derivable from `status_code` and `error_message` for every row.
- `error_message` MUST pass UI redaction before serialization.
- `actor_id` MUST equal `api_key_id` for request-action rows.

## Audit Filter (query input)

Inputs accepted by `GET /v1/admin/audit`. All fields are optional. Filter combinations use AND semantics.

- `actor`: actor identity filter. Matches `audit_log.api_key_id` for request rows.
- `api_key_id`: deprecated alias of `actor`. When both are present, `actor` wins and the handler returns 400.
- `action`: action classifier filter. Accepts a family prefix (`request:`) or an exact action (`request:gpt-4o`). The handler translates an exact `request:{model}` into a `model = $n` predicate; family prefixes resolve to action-family predicates that match all rows of that family.
- `model`: existing column-level model filter. Coexists with `action`; if both are set, both must match (intersection).
- `provider`: existing provider filter.
- `status`: legacy outcome bucket (`success` or `error`). Continues to work; when both `status` and `outcome` are present, `outcome` wins.
- `outcome`: UI-aligned outcome filter (`success`, `client_error`, `unauthorized`, `rate_limited`, `budget_denied`, `failure`). Each outcome maps to a deterministic SQL predicate over `status_code` (and, for `budget_denied`, also `error_message LIKE`).
- `from`, `to`: ISO 8601 timestamps (`time.RFC3339`). `from <= to` MUST hold.
- `limit`: 1..1000 (default 100).
- `offset`: ≥ 0 (default 0).

Validation rules:

- Unknown `outcome` values MUST return `invalid_request_error` with a list of accepted values.
- Malformed `from` or `to` MUST return `invalid_request_error`.
- `from > to` MUST return `invalid_request_error`.
- `limit ≤ 0` or non-numeric MUST return `invalid_request_error`.
- `offset < 0` or non-numeric MUST return `invalid_request_error`.
- Both `actor` and `api_key_id` set MUST return `invalid_request_error` (`code: "ambiguous_actor"`).

## Audit Page Result (response envelope)

- `data`: ordered `AuditRecord[]`, newest-first by `created_at DESC, id DESC`.
- `total`: total matches across all pages for the active filter (`int`).
- `limit`: echoed effective limit (`int`).
- `offset`: echoed effective offset (`int`).
- `page`: 1-indexed page derived from `offset/limit + 1` (`int`).
- `has_more`: `true` when `offset + len(data) < total` (`bool`).

Validation rules:

- For any filter, paginating `offset = k * limit` for `k = 0..n` MUST produce disjoint, contiguous slices of the same total result set with no duplicates and no gaps (covers SC-002).

## Reference Summary

Optional row-level annotation when an `actor_id`, `model`, or `provider` reference cannot be resolved to a known entity in the live admin tables.

- `field`: `actor_id`, `model`, or `provider` (`string`).
- `original_id`: the value carried on the audit row (`string`, may be empty).
- `state`: `missing` (empty/zero on the row), `deleted` (resolvable to a tombstoned record — reserved for future use when the joins are added), or `unknown` (non-empty on the row but not resolvable cheaply).

Validation rules:

- `field` MUST be one of the three named values.
- `original_id` MUST be the exact value from the row (no redaction); audit references are not secret-bearing.
- `state = "deleted"` MUST be backed by a tombstone lookup; until the join is implemented, only `missing` and `unknown` are emitted.

## Outcome Derivation Rules

Deterministic mapping from `(status_code, error_message)` to `outcome`:

| Status code         | error_message contains "budget"  | outcome         |
|---------------------|----------------------------------|-----------------|
| `200`–`399`         | (any)                            | `success`       |
| `401`, `403`        | (any)                            | `unauthorized`  |
| `429`               | yes                              | `budget_denied` |
| `429`               | no                               | `rate_limited`  |
| `400`–`499` (other) | (any)                            | `client_error`  |
| `500`–`599`         | (any)                            | `failure`       |

The match for `budget_denied` is a case-insensitive substring search anchored to the keyword `budget`. If the error message has been redacted, the redaction substitution MUST preserve the keyword (the redactor only removes credential-like substrings, not the surrounding human-readable text).
