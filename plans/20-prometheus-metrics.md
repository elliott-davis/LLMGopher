# Spec 20: Prometheus Metrics Endpoint

## Status
pending

## Goal
Expose a `/metrics` endpoint in Prometheus text format covering request counts, latency, token usage, cost, and error rates. This provides a production-grade observability surface compatible with Grafana, alerting rules, and any Prometheus-compatible scraper.

## Background
`go.opentelemetry.io/otel/metric` is already an indirect dependency. However, Prometheus client is more direct and battle-tested for this use case. Add `github.com/prometheus/client_golang` directly.

The existing slog logging in middleware and the cost worker provides the raw data; this spec instruments the same paths with metric counters/histograms.

## Requirements

### 1. Add dependency

```
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### 2. Metrics registry (`internal/metrics/metrics.go`)

Define all metrics using `promauto` (auto-registers with the default registry):

```go
var (
    // HTTP layer
    RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_requests_total",
        Help: "Total number of requests by model, provider, and status",
    }, []string{"model", "provider", "status_code"})

    RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "llmgopher_request_duration_seconds",
        Help:    "Request latency in seconds",
        Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5, 10, 30},
    }, []string{"model", "provider", "streaming"})

    // Token usage
    PromptTokensTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_prompt_tokens_total",
        Help: "Total prompt tokens consumed",
    }, []string{"model", "provider"})

    CompletionTokensTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_completion_tokens_total",
        Help: "Total completion tokens generated",
    }, []string{"model", "provider"})

    // Cost
    CostUSDTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_cost_usd_total",
        Help: "Total cost in USD",
    }, []string{"model", "provider"})

    // Rate limiting
    RateLimitHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_rate_limit_hits_total",
        Help: "Number of requests rejected by rate limiter",
    }, []string{"key_type"}) // "key" or "model"

    // Cache (if spec 19 is implemented)
    CacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_cache_hits_total",
        Help: "Number of cache hits",
    }, []string{"model"})

    CacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_cache_misses_total",
        Help: "Number of cache misses",
    }, []string{"model"})

    // Errors
    ProviderErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "llmgopher_provider_errors_total",
        Help: "Provider errors by type",
    }, []string{"provider", "error_type"}) // error_type: "5xx", "4xx", "timeout", "connection"

    // Budget
    BudgetUtilization = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "llmgopher_budget_utilization_ratio",
        Help: "Ratio of spent to total budget per API key (0-1)",
    }, []string{"api_key_id"})
)
```

### 3. Instrumentation points

**`internal/proxy/cost_worker.go`** — after recording a completed request, update:
- `RequestsTotal.WithLabelValues(model, provider, strconv.Itoa(statusCode)).Inc()`
- `RequestDuration.WithLabelValues(model, provider, streamingStr).Observe(latencySeconds)`
- `PromptTokensTotal`, `CompletionTokensTotal`, `CostUSDTotal`

**`internal/middleware/ratelimit.go`** — on 429:
- `RateLimitHitsTotal.WithLabelValues("key").Inc()`

**`internal/proxy/handler.go`** — on model rate limit hit (spec 09):
- `RateLimitHitsTotal.WithLabelValues("model").Inc()`

**`internal/proxy/handler.go`** — on cache hit/miss (spec 19):
- `CacheHitsTotal` / `CacheMissesTotal`

**`internal/proxy/retry.go`** — on provider error (per attempt):
- `ProviderErrorsTotal.WithLabelValues(provider, errorType).Inc()`

**`internal/storage/budget_tracker.go`** — update `BudgetUtilization` gauge after each deduction.

### 4. Metrics endpoint

In `internal/api/router.go`, add:
```go
mux.Handle("GET /metrics", promhttp.Handler())
```

This endpoint is **unauthenticated** (standard Prometheus convention; scraping is typically network-restricted). Add a note in config docs that this should be firewall-protected in production.

### 5. Build tags / conditional compilation

The metrics package uses global `promauto` registration. If downstream code avoids importing `internal/metrics`, the counters don't exist. Ensure `internal/metrics` is imported in `cmd/gateway/main.go` via a blank import if not already referenced directly.

## Out of Scope
- OpenTelemetry metrics export (spec 21)
- Per-user or per-team metrics (too high cardinality for Prometheus labels)
- Dashboard provisioning (a companion Grafana dashboard JSON can be added as a separate file)

## Acceptance Criteria
- [ ] `GET /metrics` returns valid Prometheus text format
- [ ] `llmgopher_requests_total` increments correctly by model/provider/status
- [ ] `llmgopher_request_duration_seconds` bucket counts are non-zero after requests
- [ ] `llmgopher_prompt_tokens_total` and `llmgopher_completion_tokens_total` match audit log values
- [ ] `llmgopher_rate_limit_hits_total` increments on 429 responses
- [ ] `llmgopher_cost_usd_total` matches sum of `cost_usd` in audit log
- [ ] Endpoint requires no auth (scraper-friendly)
- [ ] Metrics survive gateway restart (counters start from 0 on restart — this is expected for Prometheus counters)

## Key Files
- `internal/metrics/metrics.go` — new file, all metric definitions
- `internal/proxy/cost_worker.go` — record metrics after request
- `internal/middleware/ratelimit.go` — rate limit hit counter
- `internal/proxy/handler.go` — cache hit/miss, model rate limit
- `internal/api/router.go` — `/metrics` route
- `cmd/gateway/main.go` — blank import of `internal/metrics`
