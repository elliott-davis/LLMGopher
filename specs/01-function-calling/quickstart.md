# Quickstart: Function Calling / Tool Use Verification

## Prerequisites

- Gateway can run locally with configured providers.
- At least one configured model supports tool use.
- Local dev API key is available, for example `sk-test-key-1:key-001` when using the dev stack.

## Automated Checks

```bash
go test ./pkg/llm ./internal/proxy ./internal/providers/google/...
```

## Functional Smoke Test

1. Start the gateway with the normal local development environment.
2. Send a chat completion request containing one function tool and `tool_choice: "auto"`.
3. Verify the response can include `choices[0].message.tool_calls`.
4. Send the follow-up request with a `role: "tool"` message containing the tool result.
5. Repeat with streaming enabled and verify argument fragments arrive in ordered `tool_calls` deltas.

## Completion Signal

Mark this feature functionally verified only after automated tests pass and at least one running-gateway tool call round trip has been recorded.
