# Quickstart: GET /v1/models Verification

## Prerequisites

- Gateway can run locally with at least one configured model.
- Local dev API key is available, for example `sk-test-key-1:key-001` when using the dev stack.

## Automated Checks

```bash
go test ./internal/api/...
```

## Functional Smoke Test

1. Start the gateway with the normal local development environment.
2. Call `GET /v1/models` with a valid API key.
3. Verify the response has `object: "list"` and `data` entries with `id`, `object`, `created`, and `owned_by`.
4. Use one returned `id` as the model in a chat completion request.
5. Call the endpoint without an API key and verify the standard authentication error.
6. If possible, run an OpenAI SDK `models.list()` call against the gateway.

## Completion Signal

Mark this feature functionally verified only after automated tests pass and a running-gateway or SDK `models.list()` smoke test has been recorded.
