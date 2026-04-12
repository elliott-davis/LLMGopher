# Spec 22: Observability Callbacks (Webhooks, Langfuse, LangSmith, Helicone)

## Status
pending

## Goal
Add a pluggable callback system that fires on request success and failure, enabling integration with LLM observability platforms (Langfuse, LangSmith, Helicone) and generic webhooks. This unlocks traces, prompt management, and evals without changing client code.

## Background
`internal/proxy/cost_worker.go` already processes completed request data asynchronously via a channel. This pattern is the right foundation for callbacks — they run in the same async worker, after the response is sent to the client.

## Requirements

### 1. Callback interface (`pkg/llm/callback.go`)

```go
// CallbackEvent is fired after each request completes (success or failure).
type CallbackEvent struct {
    RequestID   string
    APIKeyID    string
    Model       string
    Provider    string
    Messages    []Message       // input messages
    Response    *ChatCompletionResponse // nil on failure
    Error       error           // nil on success
    PromptTokens     int
    CompletionTokens int
    CostUSD     float64
    LatencyMS   int64
    Streaming   bool
    Metadata    map[string]string // from API key metadata
}

// Callback is called asynchronously after each request.
type Callback interface {
    Name() string
    OnSuccess(ctx context.Context, event *CallbackEvent) error
    OnFailure(ctx context.Context, event *CallbackEvent) error
}
```

### 2. Callback registry (`internal/proxy/cost_worker.go`)

Add `[]llm.Callback` to `CostWorker`. After recording audit log and budget deduction:
```go
for _, cb := range w.callbacks {
    if event.Error == nil {
        go func(cb llm.Callback) {
            if err := cb.OnSuccess(ctx, event); err != nil {
                w.logger.Warn("callback error", "callback", cb.Name(), "error", err)
            }
        }(cb)
    } else {
        // similar for OnFailure
    }
}
```

Callbacks run in separate goroutines with a 30-second timeout context. Errors are logged but never affect the request path.

### 3. Generic webhook callback (`internal/callbacks/webhook.go`)

```go
type WebhookCallback struct {
    URL     string
    Headers map[string]string // e.g., Authorization header
    client  *http.Client
}
```

Sends a POST with JSON body:
```json
{
  "event": "success",
  "request_id": "...",
  "model": "gpt-4o",
  "provider": "openai",
  "prompt_tokens": 100,
  "completion_tokens": 50,
  "cost_usd": 0.00225,
  "latency_ms": 823,
  "timestamp": "2025-01-01T12:00:00Z"
}
```

Includes the full messages array only if `include_messages: true` is configured (off by default for privacy).

### 4. Langfuse callback (`internal/callbacks/langfuse.go`)

[Langfuse](https://langfuse.com) ingestion API (`POST https://cloud.langfuse.com/api/public/ingestion`).

Translates `CallbackEvent` to a Langfuse `generation` event:
```json
{
  "batch": [{
    "type": "generation-create",
    "body": {
      "traceId": "<request_id>",
      "name": "<model>",
      "model": "<model>",
      "input": [{"role": "user", "content": "..."}],
      "output": {"role": "assistant", "content": "..."},
      "usage": {"input": 100, "output": 50},
      "metadata": {}
    }
  }]
}
```

Auth: Basic auth with `public_key:secret_key`.

Config:
```go
Langfuse struct {
    PublicKey string `mapstructure:"public_key"`
    SecretKey string `mapstructure:"secret_key"`
    Host      string `mapstructure:"host"` // default "https://cloud.langfuse.com"
} `mapstructure:"langfuse"`
```

### 5. LangSmith callback (`internal/callbacks/langsmith.go`)

[LangSmith](https://smith.langchain.com) REST API (`POST https://api.smith.langchain.com/runs`).

Translates to a LangSmith `llm` run:
```json
{
  "name": "<model>",
  "run_type": "llm",
  "inputs": {"messages": [...]},
  "outputs": {"generations": [{"text": "..."}]},
  "extra": {"invocation_params": {"model": "gpt-4o"}},
  "start_time": "...",
  "end_time": "..."
}
```

Auth: `x-api-key` header.

Config:
```go
LangSmith struct {
    APIKey  string `mapstructure:"api_key"`
    Project string `mapstructure:"project"` // default "llmgopher"
} `mapstructure:"langsmith"`
```

### 6. Helicone callback (`internal/callbacks/helicone.go`)

Helicone uses a header-based proxy approach but also has a [logging API](https://docs.helicone.ai/references/api-reference/request/log-request). Use the logging API:

`POST https://api.hconeai.com/api/request`

Config:
```go
Helicone struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"helicone"`
```

### 7. Config and registration (`pkg/config/config.go` + `cmd/gateway/main.go`)

```go
Callbacks struct {
    Webhooks  []WebhookConfig  `mapstructure:"webhooks"`
    Langfuse  LangfuseConfig   `mapstructure:"langfuse"`
    LangSmith LangSmithConfig  `mapstructure:"langsmith"`
    Helicone  HeliconeConfig   `mapstructure:"helicone"`
} `mapstructure:"callbacks"`
```

In `cmd/gateway/main.go`, instantiate enabled callbacks and pass them to `CostWorker`.

### 8. Messages inclusion control

By default, `messages` and `response` content are NOT sent to callbacks (privacy). Add a per-callback `include_messages: true` option. When false, the callback event has nil/empty messages.

## Out of Scope
- Synchronous callbacks (all are async fire-and-forget)
- Callback retries (log and discard on error)
- Custom callback code execution (webhooks only, no arbitrary code)
- Real-time streaming callbacks

## Acceptance Criteria
- [ ] A configured webhook receives a POST after each request completes
- [ ] Langfuse receives a generation event visible in the Langfuse UI
- [ ] LangSmith receives a run event visible in the LangSmith project
- [ ] Callback errors do not affect request latency or response
- [ ] `include_messages: false` (default) sends no prompt/response content to external services
- [ ] Multiple webhooks can be configured simultaneously
- [ ] Callbacks are initialized only when their config keys are non-empty

## Key Files
- `pkg/llm/callback.go` — `Callback` interface, `CallbackEvent` type (new file)
- `internal/callbacks/webhook.go` — webhook implementation (new file)
- `internal/callbacks/langfuse.go` — Langfuse implementation (new file)
- `internal/callbacks/langsmith.go` — LangSmith implementation (new file)
- `internal/callbacks/helicone.go` — Helicone implementation (new file)
- `internal/proxy/cost_worker.go` — add callback dispatch
- `pkg/config/config.go` — callback config sections
- `cmd/gateway/main.go` — register callbacks
