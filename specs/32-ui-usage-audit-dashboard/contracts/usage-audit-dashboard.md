# Contract: Usage and Audit Dashboard

This contract documents the backend APIs consumed by the admin UI. All requests are issued by the UI server process and include:

```http
Authorization: Bearer ${LLMGOPHER_UI_ADMIN_API_KEY}
```

The token must never be sent to client-side browser code.

## GET /v1/admin/usage

Returns grouped usage and spend summaries.

**Query parameters**:
- `group_by` (required): `model`, `provider`, or `api_key`.
- `from` (optional): ISO 8601 timestamp.
- `to` (optional): ISO 8601 timestamp.
- `api_key_id` (optional): API key ID filter.
- `model` (optional): model filter.

**Success response**:

```json
{
  "group_by": "model",
  "from": "2026-04-01T00:00:00Z",
  "to": "2026-05-01T00:00:00Z",
  "data": [
    {
      "group": "gpt-4o",
      "requests": 42,
      "prompt_tokens": 12000,
      "completion_tokens": 8000,
      "total_tokens": 20000,
      "cost_usd": 1.23,
      "errors": 2,
      "avg_latency_ms": 845.5
    }
  ]
}
```

**Error handling**:
- `400`: Invalid `group_by`, timestamp, or time window. Render invalid-filter state.
- `401`/`403`: Missing or rejected admin token. Render unavailable auth guidance.
- `503`: Database unavailable. Render unavailable backend guidance.

## GET /v1/admin/usage/daily

Returns daily usage trend points.

**Query parameters**:
- `from` (optional): ISO 8601 timestamp.
- `to` (optional): ISO 8601 timestamp.
- `api_key_id` (optional): API key ID filter.
- `model` (optional): model filter.

**Success response**:

```json
{
  "from": "2026-04-01T00:00:00Z",
  "to": "2026-05-01T00:00:00Z",
  "data": [
    {
      "date": "2026-04-01",
      "requests": 12,
      "total_tokens": 4900,
      "cost_usd": 0.42
    }
  ]
}
```

**Error handling**:
- Same mapping as `GET /v1/admin/usage`.

## GET /v1/admin/audit

Returns paginated audit records with total count.

**Query parameters**:
- `api_key_id` (optional): API key ID filter.
- `model` (optional): model filter.
- `provider` (optional): provider filter.
- `status` (optional): `success` or `error`.
- `from` (optional): ISO 8601 timestamp.
- `to` (optional): ISO 8601 timestamp.
- `limit` (optional): positive integer, capped by backend at 1000.
- `offset` (optional): non-negative integer.

**Success response**:

```json
{
  "data": [
    {
      "id": 123,
      "request_id": "req-abc",
      "api_key_id": "key-001",
      "model": "gpt-4o",
      "provider": "openai",
      "prompt_tokens": 100,
      "output_tokens": 50,
      "total_tokens": 150,
      "cost_usd": 0.01,
      "status_code": 200,
      "latency_ms": 742,
      "streaming": false,
      "error_message": "",
      "created_at": "2026-05-01T12:34:56Z"
    }
  ],
  "total": 1,
  "limit": 100,
  "offset": 0
}
```

**Error handling**:
- `400`: Invalid status, timestamp, limit, offset, or time window. Render invalid-filter state.
- `401`/`403`: Missing or rejected admin token. Render unavailable auth guidance.
- `503`: Database unavailable. Render unavailable backend guidance.

## UI Contract

The dashboard adds:
- A navigation item named `Usage & Audit` linking to `/usage`.
- A dashboard card linking to `/usage`.
- A `/usage` admin page with grouped usage summary, daily trend, and audit search sections.

The page must:
- Preserve selected filters in the URL.
- Reset audit `offset` to `0` when filters other than pagination change.
- Render empty states for successful zero-row responses.
- Render unavailable states without clearing filters.
- Avoid raw API keys, provider credentials, or request payloads.
