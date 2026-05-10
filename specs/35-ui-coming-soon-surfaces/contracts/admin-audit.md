# Contract: Admin Audit Surface

## Endpoint

- `GET /v1/admin/audit`

The production audit response from `32-ui-usage-audit-dashboard` is the source of truth. The UI may use the existing E2E mock contract while keeping visible fields compatible with production audit data.

## Query Parameters

- `actor`: optional actor/API-key filter.
- `action`: optional action or model-like filter.
- `from`: optional ISO date.
- `to`: optional ISO date.
- `page`: optional positive integer.
- `limit`: optional positive integer.

Active filters must be encoded in the browser URL query string and preserved across pagination.

## Response

```json
{
  "data": [
    {
      "id": "audit_001",
      "request_id": "req_001",
      "actor_id": "key_001",
      "action": "chat.completion",
      "api_key_id": "key_001",
      "model": "gpt-4o-mini",
      "provider": "openai",
      "prompt_tokens": 100,
      "output_tokens": 50,
      "total_tokens": 150,
      "cost_usd": 0.003,
      "status_code": 200,
      "latency_ms": 900,
      "streaming": false,
      "error_message": null,
      "created_at": "2026-05-09T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 10
  }
}
```

## UI Rules

- Audit records are read-only and sorted newest-first.
- Successful, failed, unauthorized, and budget-denied outcomes must use text labels, not color alone.
- Error messages and metadata must be redacted before render.
