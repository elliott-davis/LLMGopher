# Contract: Admin Logs Surface

## Endpoints

- `GET /v1/admin/logs`
- `GET /v1/admin/logs/{id}`

The first implementation may use the existing mock contract from `specs/34-ui-e2e-testing-suite/contracts/admin-logs.md`. Production enablement requires reconciliation with the real gateway admin log API.

## Query Parameters

- `status`: optional `all`, `2xx`, `4xx`, `5xx`, or `fallback`.
- `page`: optional positive integer.
- `limit`: optional positive integer.

Active filters must also be represented in the browser URL query string.

## List Response

```json
{
  "data": [
    {
      "id": "log_fallback",
      "request_id": "req_001",
      "timestamp": "2026-05-09T00:00:00Z",
      "method": "POST",
      "path": "/v1/chat/completions",
      "status_code": 200,
      "latency_ms": 1234,
      "api_key_id": "key_001",
      "model": "gpt-4o-mini",
      "provider_chain": [
        {
          "provider_id": "openai-primary",
          "provider_name": "OpenAI Primary",
          "outcome": "failed",
          "latency_ms": 640,
          "error_summary": "timeout"
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 20
  }
}
```

## Detail Response

```json
{
  "id": "log_fallback",
  "prompt_preview": "Summarize...",
  "response_preview": "Summary...",
  "headers": {
    "authorization": "[REDACTED]"
  },
  "trace": []
}
```

## UI Rules

- Authorization headers, provider credentials, raw API keys, tokens, cookies, and secret-like values render as `[REDACTED]`.
- Prompt and response values render as redacted, truncated previews only.
- A fallback-filtered row must have a multi-stage provider chain with a failed primary stage.
