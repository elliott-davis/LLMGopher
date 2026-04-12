# Spec 06: Audit Log Query API

## Status
completed

## Goal
Expose the audit log data that is already persisted in PostgreSQL via a queryable admin endpoint. Operators need this to investigate usage, debug issues, and review request history without direct database access.

## Background
`internal/storage/audit_logger.go` writes every request to the `audit_log` table with fields: `id`, `request_id`, `api_key_id`, `model`, `provider`, `prompt_tokens`, `output_tokens`, `total_tokens`, `cost_usd`, `status_code`, `latency_ms`, `streaming`, `error_message`, `created_at`. There is no read path — logs are write-only from the application's perspective.

The DB connection is already available in the handler dependencies via `internal/api/router.go`.

## Requirements

### 1. New endpoint: `GET /v1/admin/audit`

Query parameters (all optional):
- `api_key_id` — filter by key UUID
- `model` — exact match on model name
- `provider` — exact match on provider name
- `status` — `"success"` (status_code < 400) or `"error"` (status_code >= 400)
- `from` — ISO 8601 timestamp, lower bound on `created_at`
- `to` — ISO 8601 timestamp, upper bound on `created_at`
- `limit` — max rows to return, default 100, max 1000
- `offset` — for pagination, default 0

Response body:
```json
{
  "data": [
    {
      "id": 1234,
      "request_id": "...",
      "api_key_id": "...",
      "model": "gpt-4o",
      "provider": "openai",
      "prompt_tokens": 100,
      "output_tokens": 50,
      "total_tokens": 150,
      "cost_usd": 0.00225,
      "status_code": 200,
      "latency_ms": 823,
      "streaming": false,
      "error_message": "",
      "created_at": "2025-01-01T12:00:00Z"
    }
  ],
  "total": 4521,
  "limit": 100,
  "offset": 0
}
```

The `total` count is a `COUNT(*)` with the same filters (no separate query needed if using a window function or two queries).

### 2. Storage layer (`internal/storage/audit_logger.go` or new file)

Add a `QueryAuditLog` function:
```go
type AuditQuery struct {
    APIKeyID  string
    Model     string
    Provider  string
    Status    string // "success" | "error" | ""
    From      *time.Time
    To        *time.Time
    Limit     int
    Offset    int
}

type AuditQueryResult struct {
    Data  []*llm.AuditEntry
    Total int
}

func QueryAuditLog(ctx context.Context, db *sql.DB, q AuditQuery) (*AuditQueryResult, error)
```

Build the WHERE clause dynamically based on non-zero query fields. Use `$N` positional parameters. Cap `Limit` at 1000 server-side regardless of the request value.

### 3. Route and handler

Add `GET /v1/admin/audit` to `internal/api/router.go`, wired to a new handler in `internal/api/admin.go`. The handler passes the `*sql.DB` reference (already available in the handler context).

The endpoint is admin-only — require the same Bearer auth as all other admin endpoints.

### 4. Extend `llm.AuditEntry` (`pkg/llm/audit.go`)
Add `ID int64` field so the response includes the DB row ID for pagination reference.

## Out of Scope
- Log streaming / real-time tailing
- Aggregated analytics (covered in spec 08)
- Log retention / deletion policies
- Sensitive field redaction (error messages may contain provider error details)

## Acceptance Criteria
- [x] `GET /v1/admin/audit` returns paginated rows from `audit_log`
- [x] `api_key_id` filter returns only rows for that key
- [x] `from`/`to` filters work correctly with ISO 8601 input
- [x] `status=error` returns only rows with `status_code >= 400`
- [x] `limit` is capped at 1000
- [x] `total` reflects the full count matching the filters, not just the current page
- [x] Returns 400 on invalid query parameter values (e.g., non-numeric limit)
- [x] Empty result set returns `{"data": [], "total": 0, ...}`

## Key Files
- `pkg/llm/audit.go` — add `ID` field to `AuditEntry`
- `internal/storage/audit_query.go` — add `QueryAuditLog` function
- `internal/api/admin.go` — new handler
- `internal/api/router.go` — new route
