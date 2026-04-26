# Implementation Plan: Provider Fallback Chains

**Branch**: `[16-provider-fallback]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/16-provider-fallback/spec.md`

## Summary

When a provider exhausts all retries and still fails, automatically retry the request against a configured fallback provider or model. This enables zero-client-change resilience across provider outages and capacity issues. The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: Existing gateway state where applicable  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Reliability roadmap feature converted from `plans/16-provider-fallback.md`

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
specs/16-provider-fallback/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── provider_fallback_contract.md
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

Spec 15 adds per-provider retry with backoff. This spec adds a second layer: when all retries are exhausted, the request is tried on the next provider in a fallback list. ModelConfig currently has: id, provider_id, name, alias, context_window. Fallback configuration needs to be added to the model config.

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `internal/storage/migrations/00006_model_fallbacks.sql - new migration`
- `pkg/llm/types.go - FallbackModels on ModelConfig, FallbackModel/OriginalModel on AuditEntry`
- `internal/storage/cache.go - include fallback_models in model scan`
- `internal/proxy/handler.go - fallback loop in sync and stream dispatch`
- `internal/api/admin.go - accept fallback_models in create/update`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
