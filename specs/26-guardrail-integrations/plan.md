# Implementation Plan: Guardrail Integrations

**Branch**: `[26-guardrail-integrations]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/26-guardrail-integrations/spec.md`

## Summary

Add built-in guardrail implementations for PII detection/masking (Presidio), Azure AI Content Safety, and Lakera AI. Also add response-side filtering so guardrails can inspect both prompts and completions. The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: PostgreSQL-backed gateway state and migrations where required  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Advanced Features roadmap feature converted from `plans/26-guardrail-integrations.md`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - converted plan identifies the OpenAI, LiteLLM, provider, admin, or operational behavior being matched.
- **High-throughput runtime**: PASS - hot-path impact must stay bounded and async cost/audit behavior must remain asynchronous unless explicitly required for policy decisions.
- **Typed contracts**: PASS - public request, response, provider, storage, and policy contracts should use typed Go structures or documented generated contracts.
- **Routing reliability**: PASS - routing, retry, fallback, load balancing, and provider selection behavior are central to this feature.
- **Multi-tenant spend governance**: PASS - existing key, budget, usage, and audit behavior must remain intact.
- **Observability**: PASS - request IDs, structured logs, audit context, metrics, traces, callbacks, or smoke evidence must cover meaningful runtime paths.
- **Security and config**: PASS - credentials remain encrypted, hashed, and redacted; config follows established precedence and startup validation.
- **Test discipline**: PASS - success, failure, compatibility, and smoke verification are required before verified status.

## Project Structure

### Documentation (this feature)

```text
specs/26-guardrail-integrations/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── guardrail_integrations_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/api/
internal/proxy/
internal/middleware/
internal/storage/
pkg/llm/
```

**Structure Decision**: Follow the existing gateway layer boundaries and keep public contracts, provider adapters, middleware, storage, and API handlers in their established packages.

## Converted Plan Details

### Background

pkg/llm/guardrail.go defines the Guardrail interface: Check(ctx, request) -> GuardrailVerdict. The existing NeMo implementation in internal/middleware/guardrail_nemo.go only checks requests (pre-call). The middleware calls Guardrail.Check() and blocks the request if not allowed. The current architecture doesn't support response-side filtering - the guardrail runs in middleware before the provider is called, so it never sees the response.

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `pkg/llm/guardrail.go - extend interface with CheckRequest/CheckResponse, adapter`
- `internal/middleware/guardrail.go - call CheckRequest (rename existing Check call)`
- `internal/proxy/handler.go - call CheckResponse after provider response`
- `internal/middleware/guardrail_presidio.go - new file`
- `internal/middleware/guardrail_azure_content.go - new file`
- `internal/middleware/guardrail_lakera.go - new file`
- `internal/middleware/guardrail_chain.go - new file (parallel chain runner)`
- `pkg/config/config.go - new guardrail config sections`
- `cmd/gateway/main.go - build and register chained guardrail`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
