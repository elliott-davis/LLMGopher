# Spec 08: Usage & Spend Summary API

## Status
completed

## Goal
Provide aggregated usage statistics from the audit log — total tokens, cost, and request counts grouped by model, provider, or API key over a time window. This is the primary analytics surface for operators.

## Background
All request data is in `audit_log`. Spec 06 adds a row-level query API; this spec adds aggregate summaries. The DB connection is available in handler dependencies.

## Requirements

### 1. New endpoint: `GET /v1/admin/usage`

Query parameters:
- `group_by` — required: `"model"`, `"provider"`, or `"api_key"`
- `from` — ISO 8601 timestamp, default 30 days ago
- `to` — ISO 8601 timestamp, default now
- `api_key_id` — optional filter
- `model` — optional filter

Response:
```json
{
  "group_by": "model",
  "from": "2025-01-01T00:00:00Z",
  "to": "2025-02-01T00:00:00Z",
  "data": [
    {
      "group": "gpt-4o",
      "requests": 1240,
      "prompt_tokens": 450000,
      "completion_tokens": 220000,
      "total_tokens": 670000,
      "cost_usd": 15.23,
      "errors": 12,
      "avg_latency_ms": 834
    }
  ]
}
```

The SQL is a single GROUP BY query on `audit_log`. Example for `group_by=model`:
```sql
SELECT
    model AS grp,
    COUNT(*) AS requests,
    SUM(prompt_tokens) AS prompt_tokens,
    SUM(output_tokens) AS completion_tokens,
    SUM(total_tokens) AS total_tokens,
    SUM(cost_usd) AS cost_usd,
    SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) AS errors,
    AVG(latency_ms) AS avg_latency_ms
FROM audit_log
WHERE created_at BETWEEN $1 AND $2
  -- optional filters applied dynamically
GROUP BY model
ORDER BY cost_usd DESC
```

For `group_by=api_key`, the `group` value in the response is the `api_key_id` UUID.

### 2. New endpoint: `GET /v1/admin/usage/daily`

Returns the same metrics broken down by calendar day. Useful for trend charts in the UI.

Query parameters: same as above except `group_by` is fixed to a second dimension (e.g., `group_by=model` + daily bucketing).

Response:
```json
{
  "from": "2025-01-01T00:00:00Z",
  "to": "2025-01-08T00:00:00Z",
  "data": [
    {
      "date": "2025-01-01",
      "requests": 120,
      "total_tokens": 45000,
      "cost_usd": 1.02
    }
  ]
}
```

SQL: `DATE_TRUNC('day', created_at)` group with optional filters.

### 3. Storage layer (`internal/storage/usage_store.go` — new file)

```go
type UsageQuery struct {
    GroupBy  string // "model" | "provider" | "api_key"
    From     time.Time
    To       time.Time
    APIKeyID string
    Model    string
}

type UsageSummary struct {
    Group            string
    Requests         int64
    PromptTokens     int64
    CompletionTokens int64
    TotalTokens      int64
    CostUSD          float64
    Errors           int64
    AvgLatencyMS     float64
}

type DailySummary struct {
    Date         string
    Requests     int64
    TotalTokens  int64
    CostUSD      float64
}

func QueryUsage(ctx context.Context, db *sql.DB, q UsageQuery) ([]UsageSummary, error)
func QueryDailyUsage(ctx context.Context, db *sql.DB, from, to time.Time, apiKeyID, model string) ([]DailySummary, error)
```

### 4. Routes

Add `GET /v1/admin/usage` and `GET /v1/admin/usage/daily` to `internal/api/router.go`.

## Out of Scope
- Real-time streaming metrics (use Prometheus for that, spec 20)
- Per-user or per-team breakdowns (requires teams, spec 23)
- Export to CSV or external analytics systems

## Acceptance Criteria
- [x] `GET /v1/admin/usage?group_by=model` returns aggregated data per model
- [x] `GET /v1/admin/usage?group_by=provider` returns aggregated data per provider
- [x] `GET /v1/admin/usage?group_by=api_key` returns aggregated data per key
- [x] `from`/`to` filters correctly bound the time window
- [x] `api_key_id` filter limits results to one key's data
- [x] `GET /v1/admin/usage/daily` returns one entry per calendar day
- [x] Requests with `status_code >= 400` are counted in `errors`
- [x] Missing `group_by` parameter returns 400
- [x] Invalid `group_by` value returns 400

## Key Files
- `internal/storage/usage_store.go` — new file with query functions
- `internal/api/admin_usage.go` — usage handlers
- `internal/api/router.go` — new routes
