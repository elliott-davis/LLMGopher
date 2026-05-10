# Contract: Admin Routes Surface

## Endpoints

- `GET /v1/admin/routes`
- `PATCH /v1/admin/routes/{id}` only after production mutation contract reconciliation

No dedicated production or mock route contract is currently authoritative for this feature. The first implementation must define typed UI data before rendering and keep production writes unavailable until a real gateway route API exists.

## Response

```json
{
  "data": [
    {
      "id": "route_chat_default",
      "model_alias": "gpt-4o-mini",
      "strategy": "fallback",
      "enabled": true,
      "primary_provider_id": "openai-primary",
      "fallback_provider_ids": ["anthropic-secondary"],
      "health_summary": "primary degraded, fallback healthy",
      "targets": [
        {
          "provider_id": "openai-primary",
          "provider_name": "OpenAI Primary",
          "weight": 70,
          "order": 1,
          "health_state": "degraded",
          "latency_ms": 850
        }
      ]
    }
  ]
}
```

## Editable Request

```json
{
  "strategy": "weighted",
  "targets": [
    {
      "provider_id": "openai-primary",
      "weight": 70
    },
    {
      "provider_id": "anthropic-secondary",
      "weight": 30
    }
  ]
}
```

## UI Rules

- Strategy values are `single`, `fallback`, `weighted`, or `latency`.
- Weighted strategy requires non-negative weights and a total weight greater than zero.
- Fallback strategy requires a primary provider and at least one fallback provider.
- Production save controls remain disabled/unavailable until the gateway contract is reconciled.
