# Implementation Plan: POST /v1/completions Endpoint

**Branch**: `[04-completions-endpoint]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/04-completions-endpoint/spec.md`

## Summary

Add the legacy OpenAI-compatible text completions endpoint by translating supported single-string prompt requests through the existing chat completion dispatch path. The original plan includes implementation notes and checked acceptance criteria; this conversion keeps functional verification as a separate remaining step.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing chat completion handler path, middleware chain, SSE writer, cost worker, and guardrail middleware  
**Storage**: Existing audit and cost tracking storage through the cost worker  
**Testing**: `go test ./pkg/llm ./internal/api/... ./internal/middleware/...` plus OpenAI SDK completions smoke test  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Reuse chat dispatch without adding avoidable synchronous work; streaming should preserve incremental delivery  
**Constraints**: Support only single string prompts in the first version; preserve OpenAI-compatible errors  
**Scale/Scope**: Legacy text completions endpoint mapped across configured chat-capable providers

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - supports the OpenAI legacy completions subset documented in the spec.
- **High-throughput runtime**: PASS - reuses existing chat dispatch, streaming, and async cost/audit behavior.
- **Typed contracts**: PASS - completions request and response types are explicit Go contracts.
- **Routing reliability**: PASS - model resolution reuses the existing chat model dispatch path.
- **Multi-tenant spend governance**: PASS - auth, rate limit, guardrail, audit, and cost behavior match chat completions.
- **Observability**: PASS - existing request IDs, structured logs, and audit context apply.
- **Security and config**: PASS - no new credential or configuration surface.
- **Test discipline**: PASS - requires route, streaming, prompt validation, guardrail, and type round-trip tests.

## Project Structure

### Documentation (this feature)

```text
specs/04-completions-endpoint/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── openai-completions.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
pkg/llm/
└── types.go

internal/proxy/
└── handler_completions.go

internal/api/
└── router.go

internal/middleware/
└── guardrail.go
```

**Structure Decision**: Reuse the existing chat completion provider dispatch path and keep only endpoint-specific decode/encode behavior in the completions handler.

## Complexity Tracking

No constitution violations are expected.

## Verification Strategy

- Run type round-trip tests for completions request and response contracts.
- Run API tests for sync response, streaming SSE response, array prompt rejection, and audit/cost handoff.
- Run guardrail middleware tests confirming `prompt` is checked as user content.
- Run an OpenAI SDK `completions.create()` smoke test against a running gateway.
- Record smoke-test results before changing feature status from "functional verification needed" to verified.
