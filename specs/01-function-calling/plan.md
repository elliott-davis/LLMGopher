# Implementation Plan: Function Calling / Tool Use

**Branch**: `[01-function-calling]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/01-function-calling/spec.md`

## Summary

Add OpenAI-compatible function calling and tool-use support across canonical chat contracts, provider translation, streaming deltas, and token usage estimation. The original Claude plan marks implementation complete; this Spec Kit plan preserves the implementation shape and adds explicit functional verification needs.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Standard library JSON handling, existing provider clients, existing HTTP and SSE stack  
**Storage**: N/A  
**Testing**: `go test ./pkg/llm ./internal/proxy ./internal/providers/google/...` plus targeted SDK smoke tests  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: No measurable latency regression for text-only chat requests; streaming tool deltas should preserve provider event order  
**Constraints**: Preserve OpenAI-compatible error envelope, async audit/cost behavior, middleware ordering, and redaction  
**Scale/Scope**: Chat completion requests across OpenAI pass-through, Anthropic translation, and Gemini/Vertex translation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - matches OpenAI chat completion tool fields and maps provider-native tool use where supported.
- **High-throughput runtime**: PASS - tool translation is request/response-local; cost and audit remain asynchronous.
- **Typed contracts**: PASS - canonical Go request/response structs represent public tool fields.
- **Routing reliability**: N/A - no routing selection behavior changes.
- **Multi-tenant spend governance**: PASS - token estimates include tool definitions for conservative spend accounting.
- **Observability**: PASS - existing request IDs, audit context, and redaction must be preserved.
- **Security and config**: PASS - no credential handling or runtime config change.
- **Test discipline**: PASS - requires unit and streaming tests plus SDK smoke verification.

## Project Structure

### Documentation (this feature)

```text
specs/01-function-calling/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── openai-chat-tools.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
pkg/llm/
└── types.go

internal/proxy/
├── provider_openai.go
├── provider_anthropic.go
├── provider_anthropic_test.go
└── tokencount.go

internal/providers/google/
├── translate_gemini.go
└── translate_vertex.go
```

**Structure Decision**: Use the existing domain-first gateway layout: canonical contracts in `pkg/llm`, provider adapters in `internal/proxy` and `internal/providers/google`, and focused tests next to adapters.

## Complexity Tracking

No constitution violations are expected.

## Verification Strategy

- Run unit tests covering canonical JSON round trips, OpenAI pass-through, Anthropic request/response translation, Anthropic streaming deltas, and Gemini translation.
- Run a functional smoke test against a running gateway with an OpenAI-compatible SDK tool call request.
- Confirm existing text-only chat completion tests still pass.
- Record smoke-test results before changing feature status from "functional verification needed" to verified.
