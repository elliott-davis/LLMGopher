# Contract: Admin Settings Surface

## Endpoints

- `GET /v1/admin/settings` only if a production settings API is introduced
- `PATCH /v1/admin/settings/{card_id}` only after production mutation contract reconciliation

No production settings contract is defined for this feature. The first implementation starts with safe read-only cards and local-only display preferences when they can be persisted safely in the browser.

## Card Model

```json
{
  "data": [
    {
      "id": "gateway-profile",
      "title": "Gateway Profile",
      "description": "Safe gateway identity and environment details",
      "availability": "read_only",
      "save_capability": "none",
      "last_saved_at": null,
      "fields": [
        {
          "id": "gateway_name",
          "label": "Gateway name",
          "value": "LLMGopher",
          "input_type": "text",
          "read_only": true,
          "validation_message": null
        }
      ]
    }
  ]
}
```

## Required Cards

- Gateway Profile
- Security
- Notifications
- Display

## UI Rules

- Cards without a backing API render read-only or unavailable with explanatory copy.
- Provider credential keys, admin API keys, raw config file values, and secret-like values are omitted or redacted.
- Editable cards must expose dirty, saving, success, validation error, failure, and unavailable states.
- Display preferences may be local-only if that is explicit in the UI copy and implementation.
