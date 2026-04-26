# Implementation Plan: OpenTelemetry Distributed Tracing

**Branch**: `[21-opentelemetry]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/21-opentelemetry/spec.md`

## Summary

Add distributed tracing using OpenTelemetry so that every gateway request produces a trace spanning the middleware chain, provider call, and async cost recording. Traces export to any OTEL-compatible backend (Jaeger, Datadog, Honeycomb, Grafana Tempo, etc.). The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: Existing gateway state where applicable  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Observability roadmap feature converted from `plans/21-opentelemetry.md`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - converted plan identifies the OpenAI, LiteLLM, provider, admin, or operational behavior being matched.
- **High-throughput runtime**: PASS - hot-path impact must stay bounded and async cost/audit behavior must remain asynchronous unless explicitly required for policy decisions.
- **Typed contracts**: PASS - public request, response, provider, storage, and policy contracts should use typed Go structures or documented generated contracts.
- **Routing reliability**: N/A - no first-order routing behavior changes unless introduced by implementation details.
- **Multi-tenant spend governance**: PASS - existing key, budget, usage, and audit behavior must remain intact.
- **Observability**: PASS - request IDs, structured logs, audit context, metrics, traces, callbacks, or smoke evidence must cover meaningful runtime paths.
- **Security and config**: PASS - credentials remain encrypted, hashed, and redacted; config follows established precedence and startup validation.
- **Test discipline**: PASS - success, failure, compatibility, and smoke verification are required before verified status.

## Project Structure

### Documentation (this feature)

```text
specs/21-opentelemetry/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── opentelemetry_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/api/
internal/proxy/
internal/middleware/
internal/storage/
```

**Structure Decision**: Follow the existing gateway layer boundaries and keep public contracts, provider adapters, middleware, storage, and API handlers in their established packages.

## Converted Plan Details

### Background

go.opentelemetry.io/otel, go.opentelemetry.io/otel/trace, and go.opentelemetry.io/otel/metric are already indirect dependencies (pulled in by Google Cloud SDKs). go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp is also already present. These can be promoted to direct dependencies without adding new modules.

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `internal/telemetry/tracer.go - new file, tracer initialization`
- `pkg/config/config.go - tracing config`
- `cmd/gateway/main.go - init tracer, defer shutdown`
- `internal/api/router.go - otelhttp wrapper`
- `internal/proxy/handler.go - provider call spans`
- `internal/middleware/auth.go, guardrail.go, ratelimit.go - middleware spans`
- `internal/proxy/cost_worker.go - async linked spans`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
