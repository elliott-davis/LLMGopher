# Quickstart: Vision / Image Inputs Verification

## Prerequisites

- Gateway can run locally with configured providers.
- At least one configured model accepts image inputs.
- `01-function-calling` structured message content behavior is present.

## Automated Checks

```bash
go test ./pkg/llm ./internal/proxy ./internal/providers/google/...
```

## Functional Smoke Test

1. Start the gateway with the normal local development environment.
2. Send a chat completion request whose user message content contains a text part and an HTTPS `image_url` part.
3. Verify the request succeeds and provider translation preserves the image.
4. Send a second request with a base64 `data:image/...;base64,...` image URL.
5. Verify malformed base64 input returns a clear OpenAI-compatible error.
6. Send a text-only chat request and verify behavior is unchanged.

## Completion Signal

Mark this feature functionally verified only after automated tests pass and at least one running-gateway or provider-fixture vision request has been recorded.
