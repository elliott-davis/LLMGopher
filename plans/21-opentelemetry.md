# Spec 21: OpenTelemetry Distributed Tracing

## Status
pending

## Goal
Add distributed tracing using OpenTelemetry so that every gateway request produces a trace spanning the middleware chain, provider call, and async cost recording. Traces export to any OTEL-compatible backend (Jaeger, Datadog, Honeycomb, Grafana Tempo, etc.).

## Background
`go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/trace`, and `go.opentelemetry.io/otel/metric` are already indirect dependencies (pulled in by Google Cloud SDKs). `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` is also already present. These can be promoted to direct dependencies without adding new modules.

## Requirements

### 1. New dependencies (promote from indirect)

```
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/trace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```

### 2. Tracer initialization (`internal/telemetry/tracer.go`)

```go
func InitTracer(ctx context.Context, cfg TracingConfig) (*sdktrace.TracerProvider, error)
```

`TracingConfig`:
```go
type TracingConfig struct {
    Enabled      bool
    Endpoint     string // OTLP endpoint, e.g. "http://localhost:4317" (gRPC) or "http://localhost:4318" (HTTP)
    Protocol     string // "grpc" | "http" — default "grpc"
    ServiceName  string // default "llmgopher"
    SampleRate   float64 // 0.0 to 1.0, default 1.0
}
```

If `!cfg.Enabled`, install a no-op tracer provider (traces are created but not exported — zero overhead).

The tracer provider is shut down gracefully in `cmd/gateway/main.go` during graceful shutdown.

### 3. Config (`pkg/config/config.go`)

```go
Tracing struct {
    Enabled    bool    `mapstructure:"enabled"`
    Endpoint   string  `mapstructure:"endpoint"`
    Protocol   string  `mapstructure:"protocol"`
    SampleRate float64 `mapstructure:"sample_rate"`
} `mapstructure:"tracing"`
```

Env: `LLMGOPHER_TRACING_ENABLED`, `LLMGOPHER_TRACING_ENDPOINT`, etc.

### 4. HTTP server instrumentation

Wrap the HTTP handler with `otelhttp.NewHandler` in `internal/api/router.go`:
```go
handler = otelhttp.NewHandler(handler, "llmgopher", otelhttp.WithTracerProvider(tp))
```

This creates a root span for every incoming HTTP request automatically.

### 5. Request propagation (`internal/middleware/requestid.go`)

Extract the trace ID from the OTEL span and use it as the request ID if no `X-Request-ID` was provided:
```go
spanCtx := trace.SpanFromContext(r.Context()).SpanContext()
if spanCtx.IsValid() {
    requestID = spanCtx.TraceID().String()
}
```

### 6. Provider call spans (`internal/proxy/handler.go`)

Create a child span for the provider call:
```go
ctx, span := tracer.Start(ctx, "provider.chat_completion",
    trace.WithAttributes(
        attribute.String("llm.model", req.Model),
        attribute.String("llm.provider", provider.Name()),
        attribute.Bool("llm.streaming", req.Stream),
    ),
)
defer span.End()
```

Add semantic attributes after the response:
```go
span.SetAttributes(
    attribute.Int("llm.prompt_tokens", resp.Usage.PromptTokens),
    attribute.Int("llm.completion_tokens", resp.Usage.CompletionTokens),
    attribute.Float64("llm.cost_usd", costUSD),
)
```

On error, record it on the span:
```go
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
```

### 7. Middleware spans

Add spans in key middleware:

**Auth middleware** (`internal/middleware/auth.go`): Span `"middleware.auth"` with `auth.method` attribute.

**Guardrail middleware** (`internal/middleware/guardrail.go`): Span `"middleware.guardrail"` with `guardrail.allowed` attribute.

**Rate limit middleware** (`internal/middleware/ratelimit.go`): Span `"middleware.rate_limit"` with `rate_limit.allowed` attribute.

### 8. Cost worker spans

The cost worker runs async but propagates the original request's trace context:
```go
ctx = trace.ContextWithSpan(context.Background(), originalSpan) // detached context preserving trace
```

Create a span `"cost_worker.record"` as a linked span (not child, since it's async):
```go
ctx, span := tracer.Start(ctx, "cost_worker.record",
    trace.WithLinks(trace.Link{SpanContext: originalSpanCtx}),
)
```

### 9. Trace context propagation to providers

For outbound HTTP requests to providers, inject the trace context using W3C TraceContext headers:
```go
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
```

This propagates `traceparent` and `tracestate` headers to providers that support OTEL (Anthropic, OpenAI support OTEL in some configurations).

## Out of Scope
- Metrics export via OTEL (keep Prometheus for metrics, OTEL for traces)
- Log correlation via OTEL log SDK
- Baggage propagation

## Acceptance Criteria
- [ ] With `tracing.enabled: true`, a request to `/v1/chat/completions` produces a trace visible in the configured backend
- [ ] The trace has spans for: HTTP handler, auth middleware, guardrail check, provider call, cost worker
- [ ] Provider call span has `llm.model`, `llm.provider`, `llm.prompt_tokens`, `llm.completion_tokens`
- [ ] Failed requests have spans with `status = ERROR` and the error message recorded
- [ ] With `tracing.enabled: false`, no OTEL initialization occurs and requests work normally
- [ ] Tracer provider is shut down cleanly during graceful shutdown

## Key Files
- `internal/telemetry/tracer.go` — new file, tracer initialization
- `pkg/config/config.go` — tracing config
- `cmd/gateway/main.go` — init tracer, defer shutdown
- `internal/api/router.go` — otelhttp wrapper
- `internal/proxy/handler.go` — provider call spans
- `internal/middleware/auth.go`, `guardrail.go`, `ratelimit.go` — middleware spans
- `internal/proxy/cost_worker.go` — async linked spans
