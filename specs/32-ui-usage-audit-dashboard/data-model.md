# Data Model: UI Usage and Audit Dashboard

## Usage Summary

Aggregated usage and spend for one grouping value over a selected time window.

**Fields**:
- `group`: string. Model alias/name, provider name, or API key ID depending on `group_by`.
- `requests`: integer. Count of audit rows in the group.
- `prompt_tokens`: integer. Sum of prompt/input tokens.
- `completion_tokens`: integer. Sum of completion/output tokens.
- `total_tokens`: integer. Sum of total tokens.
- `cost_usd`: number. Sum of recorded cost in USD.
- `errors`: integer. Count of rows where `status_code >= 400`.
- `avg_latency_ms`: number. Average recorded latency in milliseconds.

**Validation and display rules**:
- `group` may reference deleted historical entities and must still render.
- Token and request values are non-negative.
- Cost values must preserve small non-zero amounts with clear USD formatting.
- Rows are sorted by backend order, currently cost descending then group ascending.

## Daily Usage Point

One calendar day's usage and spend totals.

**Fields**:
- `date`: string in `YYYY-MM-DD` format.
- `requests`: integer.
- `total_tokens`: integer.
- `cost_usd`: number.

**Validation and display rules**:
- Missing days in the backend response represent zero/empty days and must not break the trend view.
- The UI may fill gaps client-side for display, but it must not imply backend data exists where it does not.

## Audit Record

Historical request-level audit row for investigation.

**Fields**:
- `id`: integer.
- `request_id`: string.
- `api_key_id`: string.
- `model`: string.
- `provider`: string.
- `prompt_tokens`: integer.
- `output_tokens`: integer.
- `total_tokens`: integer.
- `cost_usd`: number.
- `status_code`: integer.
- `latency_ms`: integer.
- `streaming`: boolean.
- `error_message`: string.
- `created_at`: ISO 8601 timestamp.

**Validation and display rules**:
- `status_code >= 400` is an error row; lower status codes are successful.
- `error_message` may be empty for successful rows.
- Request payloads, raw API keys, and provider credentials are not part of this entity and must not be displayed.
- Historical `api_key_id`, `model`, or `provider` values may be stale/deleted and should remain visible as recorded.

## Analytics Filter

URL-backed filter state applied to usage, daily trend, and audit queries.

**Fields**:
- `from`: optional ISO 8601 timestamp.
- `to`: optional ISO 8601 timestamp.
- `group_by`: one of `model`, `provider`, `api_key` for usage summaries.
- `api_key_id`: optional string.
- `model`: optional string.
- `provider`: optional string for audit records.
- `status`: optional `success` or `error` for audit records.
- `limit`: positive integer for audit page size, capped by backend at 1000.
- `offset`: non-negative integer for audit pagination.

**Validation and state transitions**:
- Empty filters use backend defaults, including the default 30-day usage window.
- `from` must be before or equal to `to` when both are present.
- Invalid filters render an actionable error while preserving the user's submitted values.
- Changing any non-pagination filter resets audit `offset` to `0`.

## Analytics Fetch State

UI state for each analytics panel.

**States**:
- `loading`: Request is pending.
- `ready`: Data loaded successfully.
- `empty`: Request succeeded but returned no rows.
- `unavailable`: Backend, database, or admin authorization is unavailable.
- `invalid-filter`: The selected filter combination is rejected locally or by the gateway.

**Rules**:
- Missing `LLMGOPHER_UI_ADMIN_API_KEY`, `401`, and `403` responses map to `unavailable` with admin-token guidance.
- `400` responses map to `invalid-filter` and preserve query parameters.
- Non-authorization `5xx` responses map to `unavailable`.
