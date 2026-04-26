# Implementation Plan: Usage & Spend Summary API

**Branch**: `[08-admin-usage-api]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/08-admin-usage-api/spec.md`

## Summary

Provide aggregated usage statistics from the audit log - total tokens, cost, and request counts grouped by model, provider, or API key over a time window. This is the primary analytics surface for operators. The original plan marks this feature complete; functional smoke verification is still required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: PostgreSQL-backed gateway state and migrations where required  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Admin & Operations roadmap feature converted from `plans/08-admin-usage-api.md`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity**: PASS - converted plan identifies the OpenAI, LiteLLM, provider, admin, or operational behavior being matched.
- **High-throughput runtime**: PASS - hot-path impact must stay bounded and async cost/audit behavior must remain asynchronous unless explicitly required for policy decisions.
- **Typed contracts**: PASS - public request, response, provider, storage, and policy contracts should use typed Go structures or documented generated contracts.
- **Routing reliability**: N/A - no first-order routing behavior changes unless introduced by implementation details.
- **Multi-tenant spend governance**: PASS - admin, budget, tenant, RBAC, usage, or audit behavior is central to this feature.
- **Observability**: PASS - request IDs, structured logs, audit context, metrics, traces, callbacks, or smoke evidence must cover meaningful runtime paths.
- **Security and config**: PASS - credentials remain encrypted, hashed, and redacted; config follows established precedence and startup validation.
- **Test discipline**: PASS - success, failure, compatibility, and smoke verification are required before verified status.

## Project Structure

### Documentation (this feature)

```text
specs/08-admin-usage-api/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── admin_usage_api_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/api/
internal/storage/
internal/middleware/
migrations/
```

**Structure Decision**: Follow the existing gateway layer boundaries and keep public contracts, provider adapters, middleware, storage, and API handlers in their established packages.

## Converted Plan Details

### Background

All request data is in audit_log. Spec 06 adds a row-level query API; this spec adds aggregate summaries. The DB connection is available in handler dependencies.

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `internal/storage/usage_store.go - new file with query functions`
- `internal/api/admin_usage.go - usage handlers`
- `internal/api/router.go - new routes`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
