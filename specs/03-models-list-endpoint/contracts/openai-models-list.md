# Contract: OpenAI-Compatible Models List

## Endpoint

`GET /v1/models`

## Authentication

Requires the same API key authentication as other `/v1` endpoints.

## Successful Response

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "created": 1234567890,
      "owned_by": "llmgopher"
    }
  ]
}
```

## Empty Response

```json
{
  "object": "list",
  "data": []
}
```

## Compatibility Notes

- `id` is the gateway model alias used by inference endpoints.
- The endpoint does not discover upstream provider models dynamically.
- `GET /v1/models/{model}` is outside this contract.
- Errors must use the gateway's OpenAI-compatible error envelope.
