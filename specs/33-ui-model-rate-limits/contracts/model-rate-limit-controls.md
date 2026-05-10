# Contract: UI Model Rate Limit Controls

## Existing Admin API Surface

The UI uses the existing admin model endpoints:

- `GET /v1/admin/models`
- `POST /v1/admin/models`
- `PUT /v1/admin/models/{id}`

No new backend endpoint is planned.

## Model Response Shape

`GET /v1/admin/models` returns model objects that include:

```json
{
  "id": "model-uuid",
  "provider_id": "provider-uuid",
  "name": "gpt-4o-2024-11-20",
  "alias": "gpt-4o",
  "context_window": 128000,
  "rate_limit_rps": 25,
  "created_at": "2026-05-03T00:00:00Z",
  "updated_at": "2026-05-03T00:00:00Z"
}
```

## Create Model Payload

`POST /v1/admin/models` payload:

```json
{
  "alias": "gpt-4o",
  "name": "gpt-4o-2024-11-20",
  "provider_id": "provider-uuid",
  "context_window": 128000,
  "rate_limit_rps": 25
}
```

Expected behavior:
- `rate_limit_rps >= 0` is accepted.
- `rate_limit_rps === 0` means no model-level limit.
- `rate_limit_rps < 0` returns an OpenAI-compatible admin error envelope with type `invalid_request_error`.
- Successful create revalidates and refreshes the `/models` UI.

## Update Model Payload

`PUT /v1/admin/models/{id}` payload:

```json
{
  "alias": "gpt-4o",
  "name": "gpt-4o-2024-11-20",
  "provider_id": "provider-uuid",
  "context_window": 128000,
  "rate_limit_rps": 0
}
```

Expected behavior:
- Full model updates preserve existing alias, name, provider, and context window semantics.
- Setting `rate_limit_rps` to `0` removes model-level throttling while preserving key-level rate limits.
- Gateway failures keep the edit form state open and display the failure reason.

## UI Contract

Create and edit forms must include:
- A numeric `rate_limit_rps` control with minimum `0`.
- Default value `0` for newly created models.
- Helper text explaining that this is a model-level requests-per-second limit separate from API key limits.
- Clear feedback for negative or otherwise invalid values.

The model inventory must include:
- A dedicated model rate limit column.
- Positive values displayed with requests-per-second units.
- Zero or missing values displayed as an explicit no model-level limit state.

## Out Of Scope

- Token-per-minute limits.
- Per-key-per-model compound limits.
- New runtime rate-limit algorithms.
- New backend persistence or migrations.
