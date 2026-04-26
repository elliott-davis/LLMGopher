# Implementation Plan: Context-Window Fallback

**Branch**: `[18-context-window-fallback]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/18-context-window-fallback/spec.md`

## Summary

When a provider rejects a request with a context-length-exceeded error, automatically retry the request with a designated larger-context model. This handles the common case where a conversation grows too long for a model's context window. The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: Existing gateway state where applicable  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Reliability roadmap feature converted from `plans/18-context-window-fallback.md`

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
specs/18-context-window-fallback/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── context_window_fallback_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/proxy/
pkg/llm/
internal/storage/
```

**Structure Decision**: Follow the existing gateway layer boundaries and keep public contracts, provider adapters, middleware, storage, and API handlers in their established packages.

## Converted Plan Details

### Background

Context-length errors are distinct from general failures: they require routing to a *different, larger* model rather than retrying the same one. Spec 15 (retry) and spec 16 (fallback) handle general failures; this spec handles the specific context_length_exceeded error case. Providers signal this with different error messages: - OpenAI: {"error": {"code": "context_length_exceeded", ...}} - Anthropic: {"error": {"type": "invalid_request_error", "message": "prompt is too long: ..."}} - Gemini: 400 with "Request contains content that exceeds token limit" - Bedrock: similar to Anthropic

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `internal/storage/migrations/00008_context_fallback.sql - new migration`
- `pkg/llm/types.go - add ContextFallbackModel to ModelConfig`
- `internal/storage/cache.go - include in model scan`
- `internal/proxy/retry.go - add isContextLengthError`
- `internal/proxy/handler.go - fallback dispatch on context error`
- `internal/api/admin.go - accept field in create/update`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
