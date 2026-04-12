# Spec 26: Guardrail Integrations

## Status
pending

## Goal
Add built-in guardrail implementations for PII detection/masking (Presidio), Azure AI Content Safety, and Lakera AI. Also add response-side filtering so guardrails can inspect both prompts and completions.

## Background
`pkg/llm/guardrail.go` defines the `Guardrail` interface: `Check(ctx, request) → GuardrailVerdict`. The existing NeMo implementation in `internal/middleware/guardrail_nemo.go` only checks requests (pre-call). The middleware calls `Guardrail.Check()` and blocks the request if not allowed.

The current architecture doesn't support response-side filtering — the guardrail runs in middleware before the provider is called, so it never sees the response.

## Requirements

### 1. Response-side guardrail support

Extend the `Guardrail` interface (`pkg/llm/guardrail.go`):
```go
type Guardrail interface {
    // CheckRequest is called before the provider call.
    CheckRequest(ctx context.Context, req *ChatCompletionRequest) GuardrailVerdict

    // CheckResponse is called after the provider returns a response.
    // Returning !Allowed causes the gateway to replace the response with an error.
    CheckResponse(ctx context.Context, req *ChatCompletionRequest, resp *ChatCompletionResponse) GuardrailVerdict
}
```

For backward compatibility, provide a `RequestOnlyGuardrail` adapter that implements the new interface using a legacy `Check` function:
```go
type RequestOnlyGuardrail struct{ fn func(context.Context, *ChatCompletionRequest) GuardrailVerdict }
func (g *RequestOnlyGuardrail) CheckRequest(...) GuardrailVerdict { return g.fn(...) }
func (g *RequestOnlyGuardrail) CheckResponse(...) GuardrailVerdict { return GuardrailVerdict{Allowed: true} }
```

Wrap the existing `NemoGuardrail` in this adapter.

### 2. Response-side enforcement

Move guardrail invocation from middleware into `internal/proxy/handler.go`. The middleware continues to call `CheckRequest` (it still has access to the body). Add response checking in `handleSync` after the provider returns:

```go
if h.guardrail != nil {
    verdict := h.guardrail.CheckResponse(ctx, &req, resp)
    if !verdict.Allowed {
        writeError(w, http.StatusForbidden, verdict.Reason, "content_policy_error")
        h.costWorker.RecordError(meta, http.StatusForbidden, "response blocked by guardrail")
        return
    }
}
```

For streaming: buffer the full response to check it, then replay it to the client if allowed. This adds latency — make it configurable. If `check_response: false` (default for streaming), skip response checking for streams.

### 3. Microsoft Presidio (`internal/middleware/guardrail_presidio.go`)

Presidio is a self-hosted PII detection service (Docker image: `mcr.microsoft.com/presidio-analyzer`).

**PII detection (request-side):**
- POST to `{presidio_url}/analyze` with request text
- Returns list of detected PII entities (type, start, end, score)
- If any entity score >= threshold: block request (or mask)

**PII masking (optional):**
- POST to `{presidio_url}/anonymize` to replace PII with placeholders before forwarding
- When masking is enabled, the request is modified in-place before provider dispatch; the request is allowed (not blocked)

Config:
```go
Presidio struct {
    AnalyzerURL   string   `mapstructure:"analyzer_url"`
    AnonymizerURL string   `mapstructure:"anonymizer_url"`
    Entities      []string `mapstructure:"entities"`       // e.g. ["PERSON", "EMAIL_ADDRESS", "PHONE_NUMBER"]
    Threshold     float64  `mapstructure:"threshold"`      // default 0.7
    Mode          string   `mapstructure:"mode"`           // "block" | "mask"
} `mapstructure:"presidio"`
```

### 4. Azure AI Content Safety (`internal/middleware/guardrail_azure_content.go`)

Azure Content Safety API (`POST https://{resource}.cognitiveservices.azure.com/contentsafety/text:analyze`).

Returns scores for Hate, Violence, SelfHarm, Sexual on a 0–6 scale. Block if any category score >= configured threshold.

Config:
```go
AzureContentSafety struct {
    Endpoint  string `mapstructure:"endpoint"`   // e.g. "https://myresource.cognitiveservices.azure.com"
    APIKey    string `mapstructure:"api_key"`
    Threshold int    `mapstructure:"threshold"`  // 0-6, default 2
} `mapstructure:"azure_content_safety"`
```

Implement `CheckRequest` only (request-side). Add `CheckResponse` that calls the same API on the response text.

### 5. Lakera AI (`internal/middleware/guardrail_lakera.go`)

Lakera Guard API (`POST https://api.lakera.ai/v2/guard`).

Detects: prompt injection, jailbreak, PII, hate speech, sexual content.

Config:
```go
Lakera struct {
    APIKey string `mapstructure:"api_key"`
} `mapstructure:"lakera"`
```

Simple implementation: POST the prompt text, check if any flagged category is above threshold, return `GuardrailVerdict`.

### 6. Guardrail chaining

Support multiple active guardrails. They run in sequence; the first to block terminates the chain.

In `cmd/gateway/main.go`:
```go
var guardrails []llm.Guardrail
if cfg.Presidio.AnalyzerURL != "" { guardrails = append(guardrails, presidioGuardrail) }
if cfg.AzureContentSafety.Endpoint != "" { guardrails = append(guardrails, azureGuardrail) }
if cfg.Lakera.APIKey != "" { guardrails = append(guardrails, lakeraGuardrail) }
if cfg.Guardrail.Endpoint != "" { guardrails = append(guardrails, nemoGuardrail) }

chainedGuardrail := guardrail.NewChain(guardrails...)
```

`NewChain` runs all guardrails in parallel (with `errgroup`) for performance and returns the first non-allowed verdict.

### 7. Config registration

Add all new guardrail configs to `pkg/config/config.go`.

## Out of Scope
- Llamaguard integration (local model inference required)
- Custom guardrail plugins (arbitrary Go code)
- Per-key guardrail configuration (global guardrails only)

## Acceptance Criteria
- [ ] Presidio in `block` mode rejects a request containing a clearly detectable email address
- [ ] Presidio in `mask` mode replaces PII before forwarding to the provider
- [ ] Azure Content Safety blocks a request with highly toxic content
- [ ] Lakera detects a prompt injection attempt and returns 403
- [ ] Multiple guardrails can be active simultaneously; first to block wins
- [ ] Response-side guardrail blocks a response containing prohibited content
- [ ] Legacy NeMo integration still works unchanged
- [ ] All guardrail check results are logged (allowed/blocked + reason)

## Key Files
- `pkg/llm/guardrail.go` — extend interface with `CheckRequest`/`CheckResponse`, adapter
- `internal/middleware/guardrail.go` — call `CheckRequest` (rename existing `Check` call)
- `internal/proxy/handler.go` — call `CheckResponse` after provider response
- `internal/middleware/guardrail_presidio.go` — new file
- `internal/middleware/guardrail_azure_content.go` — new file
- `internal/middleware/guardrail_lakera.go` — new file
- `internal/middleware/guardrail_chain.go` — new file (parallel chain runner)
- `pkg/config/config.go` — new guardrail config sections
- `cmd/gateway/main.go` — build and register chained guardrail
