# Implementation Plan: Teams & Organizations

**Branch**: `[23-teams-organizations]` | **Date**: 2026-04-26 | **Spec**: `spec.md`  
**Input**: Feature specification from `specs/23-teams-organizations/spec.md`

## Summary

Add a team and organization model so that API keys can be grouped, team-level budgets and rate limits can be enforced, and usage can be attributed at the team level. This is the foundation for multi-tenant operation. The original plan marks this feature pending; implementation and verification are both required.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Existing LLMGopher gateway packages, provider clients, middleware, storage, and test utilities as applicable  
**Storage**: PostgreSQL-backed gateway state and migrations where required  
**Testing**: `go test ./...` plus focused package tests and the smoke checks in `quickstart.md`  
**Target Platform**: Gateway service  
**Project Type**: OpenAI-compatible API gateway  
**Performance Goals**: Preserve hot-path latency and streaming behavior for unaffected requests; add benchmarks or load checks when routing, caching, provider, or middleware paths change  
**Constraints**: Preserve OpenAI-compatible errors, credential redaction, async cost/audit behavior, and configuration precedence  
**Scale/Scope**: Multi-tenancy roadmap feature converted from `plans/23-teams-organizations.md`

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
specs/23-teams-organizations/
├── spec.md
├── plan.md
├── research.md
├── quickstart.md
├── contracts/
│   └── teams_organizations_contract.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/api/
internal/middleware/
internal/storage/
migrations/
```

**Structure Decision**: Follow the existing gateway layer boundaries and keep public contracts, provider adapters, middleware, storage, and API handlers in their established packages.

## Converted Plan Details

### Background

All keys are currently in a flat namespace. Adding teams requires new DB tables, foreign keys on existing tables, and enforcement logic in the request path. This spec establishes the data model. Spec 24 adds RBAC and JWT auth on top of it.

### Dependencies

- No explicit cross-plan dependencies beyond existing gateway infrastructure.

### Key Files From Original Plan

- `internal/storage/migrations/00010_teams.sql - new migration`
- `pkg/llm/types.go - Organization, Team, TeamBudget, extend APIKeyConfig`
- `internal/storage/cache.go - load teams/orgs into state`
- `internal/storage/budget_tracker.go - team deduct/remaining methods`
- `internal/proxy/handler.go - team budget pre-check`
- `internal/proxy/cost_worker.go - team budget deduction`
- `internal/api/admin.go - org/team CRUD handlers`
- `internal/api/router.go - new routes`

## Complexity Tracking

No constitution violations are expected from the conversion. Any later implementation that adds blocking hot-path calls, unbounded work, or credential exposure must document an exception here before proceeding.

## Verification Strategy

- Run focused tests for packages touched by the original plan.
- Run `go test ./...` before marking implementation complete.
- Execute the functional smoke checks in `quickstart.md`.
- Confirm logs, audit records, metrics, traces, or callback events preserve redaction and request context where applicable.
- Record smoke-test evidence before changing status from "functional verification needed" or "pending implementation" to verified.
