# Contract: Admin Logs

Real logs API is not yet shipped. The mock implements just enough for the logs surface tests.

## `GET /v1/admin/logs?status=&limit=&cursor=`

**Response 200**:
```json
{
  "data": [
    {
      "id": "log_001",
      "request_id": "req_…",
      "timestamp": "2026-05-09T11:59:42Z",
      "method": "POST",
      "path": "/v1/chat/completions",
      "status_code": 200,
      "latency_ms": 432,
      "api_key_id": "key_checkout_service",
      "model": "gpt-4o",
      "provider_chain": [
        { "provider_id": "prov_openai_prod", "status": "ok",     "latency_ms": 410 }
      ]
    }
  ],
  "next_cursor": null
}
```

`status` filter accepts `2xx | 4xx | 5xx | fallback`.

A row in the seed has `id="log_fallback"` with `provider_chain` of length ≥ 2 where the first stage has `status="failed"`. The detail view renders this as `data-failed="true"` on `data-testid="timeline-stage-primary"`.

## `GET /v1/admin/logs/{id}`

Returns the full log including `prompt`, `response`, and `headers` blobs (used by the inspector tabs). Headers MUST be redacted — `authorization` shown as `Bearer ****`.
