# Spec 04: POST /v1/completions Endpoint

## Status
implemented

## Goal
Add the legacy OpenAI text completions endpoint (`POST /v1/completions`). While superseded by chat completions, many tools (LangChain legacy chains, some evals frameworks, older integrations) still call this endpoint.

## Background
The gateway only has `POST /v1/chat/completions`. The completions endpoint uses a different request/response shape (a `prompt` string instead of `messages` array) but can be implemented by translating to a single-turn chat completion internally.

OpenAI completions request (relevant fields):
```json
{"model": "gpt-3.5-turbo-instruct", "prompt": "Say hello", "max_tokens": 100, "stream": false}
```

OpenAI completions response:
```json
{"id": "...", "object": "text_completion", "created": 1234, "model": "...",
 "choices": [{"text": "Hello!", "index": 0, "finish_reason": "stop"}],
 "usage": {"prompt_tokens": 3, "completion_tokens": 5, "total_tokens": 8}}
```

## Requirements

1. Add new canonical types in `pkg/llm/types.go`:

```go
type CompletionRequest struct {
    Model            string          `json:"model"`
    Prompt           string          `json:"prompt"` // single string only (not array)
    MaxTokens        *int            `json:"max_tokens,omitempty"`
    Temperature      *float64        `json:"temperature,omitempty"`
    TopP             *float64        `json:"top_p,omitempty"`
    Stream           bool            `json:"stream,omitempty"`
    Stop             json.RawMessage `json:"stop,omitempty"`
    PresencePenalty  float64         `json:"presence_penalty,omitempty"`
    FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
    User             string          `json:"user,omitempty"`
}

type CompletionChoice struct {
    Text         string `json:"text"`
    Index        int    `json:"index"`
    FinishReason string `json:"finish_reason,omitempty"`
}

type CompletionResponse struct {
    ID      string             `json:"id"`
    Object  string             `json:"object"` // "text_completion"
    Created int64              `json:"created"`
    Model   string             `json:"model"`
    Choices []CompletionChoice `json:"choices"`
    Usage   *Usage             `json:"usage,omitempty"`
}
```

2. Add a handler in `internal/proxy/` that:
   - Decodes `CompletionRequest`
   - Translates it to `ChatCompletionRequest` with a single user message: `{role: "user", content: prompt}`
   - Calls the existing chat completion dispatch path (reuse `Handler.resolveModel` and provider dispatch)
   - Translates `ChatCompletionResponse` back to `CompletionResponse`: extract `choices[0].message.content` as `text`

3. For streaming: translate `ChatCompletionChunk` deltas to completions SSE format:
   - `object: "text_completion"`, `choices[0].text` = delta content

4. Wire route `POST /v1/completions` in `internal/api/router.go` with the same middleware chain as chat completions.

5. Audit logging and cost tracking apply exactly as for chat completions (reuse the same `CostWorker`).

## Out of Scope
- Prompt arrays (accept single string only; return 400 for array prompts)
- `echo`, `logprobs`, `best_of`, `suffix` parameters — accept but ignore
- Per-provider completions API (Anthropic and Gemini don't have a direct equivalent; the chat translation covers them)

## Acceptance Criteria
- [x] `POST /v1/completions` with a string prompt returns a valid completions response
- [x] Streaming completions work with SSE
- [x] `object` field in response is `"text_completion"` (not `"chat.completion"`)
- [x] Cost and audit logging records the request
- [x] Array prompt returns 400 with a clear error message
- [x] OpenAI Python SDK `client.completions.create(model=..., prompt=...)` succeeds *(protocol covered by tests; optional manual smoke test against a running gateway)*

## Implementation notes

- **Handler:** `Handler.ServeCompletionsHTTP` in `internal/proxy/handler_completions.go` decodes the body (array `prompt` rejected before JSON fully unmarshals into `CompletionRequest`), builds `ChatCompletionRequest`, reuses `resolveModel`, sync/stream provider paths, maps chat responses to `CompletionResponse` / SSE chunks with `object: "text_completion"`, and calls `CostWorker` like chat.
- **Types:** `CompletionRequest` includes optional compatibility fields (`echo`, `logprobs`, `best_of`, `suffix`) JSON-tagged for decode; routing ignores them per out-of-scope.
- **Guardrails:** `internal/middleware/guardrail.go` maps bodies with `prompt` and no `messages` to a single user message for `Guardrail.Check`, so the completions route can use the same middleware chain as chat without failing JSON decode. Invalid `prompt` shapes (e.g. array) fall through so the completions handler returns the dedicated 400 message.
- **Tests:** `internal/api/router_test.go` (sync, stream SSE, array prompt, audit); `internal/middleware/middleware_test.go` (prompt → guardrail); `pkg/llm/types_test.go` (completion types JSON round-trip).

## Key Files
- `pkg/llm/types.go` — request/response types (`CompletionRequest`, `CompletionChoice`, `CompletionResponse`)
- `internal/proxy/handler_completions.go` — completions handler
- `internal/api/router.go` — `POST /v1/completions` wired with `applyChatMiddleware`
- `internal/middleware/guardrail.go` — completions `prompt` translation for guardrail checks
