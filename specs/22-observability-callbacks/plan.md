# Implementation Plan: Observability Callbacks (Webhooks, Langfuse, LangSmith, Helicone)

**Branch**: `[22-observability-callbacks]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/22-observability-callbacks/spec.md`

## Summary

Add a pluggable callback system that fires on request success and failure, enabling integration with LLM observability platforms (Langfuse, LangSmith, Helicone) and generic webhooks. This unlocks traces, prompt management, and evals without changing client code. The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: Existing gateway state where applicable  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Observability roadmap feature converted from `plans/22-observability-callbacks.md`

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
specs/22-observability-callbacks/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── observability_callbacks_contract.md
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

internal/proxy/cost_worker.go already processes completed request data asynchronously via a channel. This pattern is the right foundation for callbacks - they run in the same async worker, after the response is sent to the client.

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `pkg/llm/callback.go - Callback interface, CallbackEvent type (new file)`
- `internal/callbacks/webhook.go - webhook implementation (new file)`
- `internal/callbacks/langfuse.go - Langfuse implementation (new file)`
- `internal/callbacks/langsmith.go - LangSmith implementation (new file)`
- `internal/callbacks/helicone.go - Helicone implementation (new file)`
- `internal/proxy/cost_worker.go - add callback dispatch`
- `pkg/config/config.go - callback config sections`
- `cmd/gateway/main.go - register callbacks`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
