# Spec 29: Batch API

## Status
pending

## Dependencies
- Spec 01 (function calling) — batch requests may include tool calls
- Spec 06 (audit log query) — batch results stored alongside regular audit entries

## Goal
Implement an OpenAI-compatible Batch API for asynchronous bulk processing of chat completion requests. Clients submit a batch of requests, receive a batch ID, poll for status, and retrieve results when complete. This enables high-throughput, cost-efficient processing of large workloads (evals, document processing, data enrichment).

## Background
OpenAI's Batch API accepts a JSONL file of requests, processes them asynchronously, and returns a JSONL file of responses. Pricing is typically 50% lower. The gateway implementation processes batches using background workers against the configured providers.

## Requirements

### 1. Database schema (`internal/storage/migrations/00013_batch_api.sql`)

```sql
CREATE TYPE batch_status AS ENUM (
    'validating', 'in_progress', 'finalizing', 'completed', 'failed', 'expired', 'cancelling', 'cancelled'
);

CREATE TABLE batches (
    id TEXT PRIMARY KEY,              -- "batch_<uuid>"
    api_key_id UUID NOT NULL REFERENCES api_keys(id),
    status batch_status NOT NULL DEFAULT 'validating',
    endpoint TEXT NOT NULL,           -- "/v1/chat/completions"
    input_file_id TEXT NOT NULL,
    output_file_id TEXT,
    error_file_id TEXT,
    request_counts_total INTEGER NOT NULL DEFAULT 0,
    request_counts_completed INTEGER NOT NULL DEFAULT 0,
    request_counts_failed INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    in_progress_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE batch_requests (
    id TEXT PRIMARY KEY,              -- "req_<uuid>"
    batch_id TEXT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    custom_id TEXT NOT NULL,          -- client-provided ID for correlation
    status TEXT NOT NULL DEFAULT 'pending', -- "pending" | "processing" | "completed" | "failed"
    method TEXT NOT NULL,
    url TEXT NOT NULL,
    body JSONB NOT NULL,
    response_status_code INTEGER,
    response_body JSONB,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE batch_files (
    id TEXT PRIMARY KEY,              -- "file_<uuid>"
    batch_id TEXT,                    -- nullable; input files are not yet associated
    filename TEXT NOT NULL,
    purpose TEXT NOT NULL,            -- "batch" | "batch_output"
    content BYTEA NOT NULL,           -- JSONL content
    size INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 2. File upload endpoint

`POST /v1/files` — accept multipart form with JSONL file:
```go
type FileUploadResponse struct {
    ID        string `json:"id"`
    Object    string `json:"object"` // "file"
    Bytes     int    `json:"bytes"`
    CreatedAt int64  `json:"created_at"`
    Filename  string `json:"filename"`
    Purpose   string `json:"purpose"`
}
```

Validate: purpose must be `"batch"`, file must be valid JSONL, each line must parse as a batch request object:
```json
{"custom_id": "req-1", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "...", "messages": [...]}}
```

Max file size: 100MB (configurable). Max 50,000 requests per file.

`GET /v1/files/{file_id}` — get file metadata.
`GET /v1/files/{file_id}/content` — download file content.
`DELETE /v1/files/{file_id}` — delete file.

### 3. Batch creation

`POST /v1/batches`:
```json
{
  "input_file_id": "file_abc123",
  "endpoint": "/v1/chat/completions",
  "completion_window": "24h",
  "metadata": {}
}
```

Response: `Batch` object with `status: "validating"`.

Validation: parse all requests in the input file, return error if any are invalid.

### 4. Batch status and retrieval

- `GET /v1/batches/{batch_id}` — get batch status
- `GET /v1/batches` — list batches (paginated, filter by status)
- `POST /v1/batches/{batch_id}/cancel` — cancel in-progress batch

### 5. Batch processing worker (`internal/batch/worker.go`)

```go
type BatchWorker struct {
    db       *sql.DB
    registry llm.ProviderRegistry
    cache    *storage.StateCache
    logger   *slog.Logger
    // concurrency controls
    sem      chan struct{} // semaphore for max concurrent requests
}

func (w *BatchWorker) Start(ctx context.Context)
func (w *BatchWorker) processNextBatch(ctx context.Context)
```

The worker polls `batches` table every 5 seconds for batches with `status = 'validating'` or `status = 'in_progress'`:

1. Transition to `in_progress`
2. Query pending `batch_requests` rows (with a configurable concurrency limit, e.g., 10 concurrent)
3. For each request: call the provider, record result in `batch_requests`
4. Update `batches.request_counts_completed/failed` after each request
5. When all requests complete, generate output JSONL, write to `batch_files`, update `status = 'completed'`

Output JSONL format per line:
```json
{"id": "req_abc", "custom_id": "req-1", "response": {"status_code": 200, "request_id": "...", "body": {...}}, "error": null}
```

### 6. Concurrency and rate limiting

Batch workers respect the same rate limits as regular requests (they use the API key's rate limit). Configure max concurrent batch requests per worker:
```go
Batch struct {
    MaxConcurrent int           `mapstructure:"max_concurrent"`  // default 10
    PollInterval  time.Duration `mapstructure:"poll_interval"`   // default 5s
    DefaultTTL    time.Duration `mapstructure:"default_ttl"`     // default 24h
} `mapstructure:"batch"`
```

### 7. Cost tracking

Each request in a batch is logged to `audit_log` individually with `batch_id` metadata. Budget deduction happens per request as normal.

Add `batch_id TEXT` to `audit_log`:
```sql
ALTER TABLE audit_log ADD COLUMN batch_id TEXT;
```

### 8. Expiry

Batches with `expires_at < NOW()` that are not yet completed transition to `status = 'expired'`. The poll worker handles this.

## Out of Scope
- Provider-native batch APIs (e.g., OpenAI's actual batch endpoint — this impl processes batches locally)
- Batch pricing discounts (cost is tracked at normal rates)
- Streaming batch results
- Webhook notification on batch completion (add via spec 22 callbacks later)

## Acceptance Criteria
- [ ] `POST /v1/files` with a valid JSONL file returns a file ID
- [ ] `POST /v1/batches` creates a batch and transitions to `in_progress`
- [ ] Batch worker processes all requests and writes output file
- [ ] `GET /v1/batches/{id}` reflects accurate `request_counts_completed`
- [ ] `GET /v1/files/{file_id}/content` returns the output JSONL when batch is complete
- [ ] `POST /v1/batches/{id}/cancel` stops processing remaining requests
- [ ] Each batch request is individually logged in `audit_log`
- [ ] Batch respects the API key's rate limit (doesn't bypass per-key throttling)
- [ ] Expired batches are marked `status: "expired"` by the worker

## Key Files
- `internal/storage/migrations/00013_batch_api.sql` — new migration
- `internal/batch/worker.go` — batch processing worker (new package)
- `internal/api/handler_files.go` — file CRUD handlers (new file)
- `internal/api/handler_batches.go` — batch CRUD handlers (new file)
- `internal/api/router.go` — new routes
- `pkg/config/config.go` — batch config
- `cmd/gateway/main.go` — start batch worker
