# Quickstart: POST /v1/completions Verification

## Prerequisites

- Gateway can run locally with at least one configured model.
- Local dev API key is available, for example `sk-test-key-1:key-001` when using the dev stack.

## Automated Checks

```bash
go test ./pkg/llm ./internal/api/... ./internal/middleware/...
```

## Functional Smoke Test

1. Start the gateway with the normal local development environment.
2. Send `POST /v1/completions` with a single string prompt and `stream: false`.
3. Verify the response object is `text_completion` and choices contain `text`.
4. Send the same request with `stream: true` and verify server-sent events contain text completion chunks.
5. Send a prompt array and verify HTTP 400 with an OpenAI-compatible error envelope.
6. If possible, run an OpenAI SDK `completions.create()` call against the gateway.

## Completion Signal

Mark this feature functionally verified only after automated tests pass and a running-gateway or SDK completions smoke test has been recorded.
