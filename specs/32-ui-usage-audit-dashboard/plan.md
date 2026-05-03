# Implementation Plan: UI Usage and Audit Dashboard

**Branch**: `[main]` | **Date**: 2026-05-03 | **Spec**: `specs/32-ui-usage-audit-dashboard/spec.md`
**Input**: Feature specification from `specs/32-ui-usage-audit-dashboard/spec.md`

## Summary

Surface the existing admin analytics APIs in the Next.js admin UI so operators can review grouped usage, daily spend/token trends, and paginated audit records without raw API calls. The implementation is UI-focused: add typed TypeScript contracts and server-side fetch helpers for `/v1/admin/usage`, `/v1/admin/usage/daily`, and `/v1/admin/audit`, wire authenticated requests through `LLMGOPHER_UI_ADMIN_API_KEY`, add dashboard navigation/cards, and build focused pages/components with loading, empty, unavailable, and invalid-filter states.

## Technical Context

**Language/Version**: Go 1.22+ backend; TypeScript with Next.js 15, React 19, and server components in `ui/`  
**Primary Dependencies**: Existing Go `net/http` admin routes, PostgreSQL-backed audit/usage queries, Next.js App Router, React, lucide-react, existing local UI primitives  
**Storage**: Existing PostgreSQL `audit_log` source of truth via `internal/storage/usage_store.go` and `internal/storage/audit_query.go`; no new persistence planned  
**Testing**: `go test ./...` for backend guardrails if touched; `npm test -- --run`/`vitest run` in `ui/` for fetch helpers, filter parsing, page rendering, pagination, and empty/error states  
**Target Platform**: Linux gateway plus Docker Compose admin UI service  
**Project Type**: Web service with companion Next.js admin UI  
**Performance Goals**: No change to request hot path; usage and audit UI fetches should be server-rendered, bounded by backend limits, and avoid client-side unbounded datasets  
**Constraints**: Do not expose API key secrets, provider credentials, or request payloads; admin analytics endpoints require a bearer token from `LLMGOPHER_UI_ADMIN_API_KEY`; preserve selected filters across unavailable and validation states  
**Scale/Scope**: One new analytics area in `ui/` with grouped usage summaries, daily trend table/cards, and audit search over backend-paginated results capped by existing API limits

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Gate

- **Upstream parity**: PASS. This feature exposes existing LLMGopher admin analytics APIs and does not alter OpenAI-compatible runtime behavior.
- **High-throughput runtime**: PASS. No request hot-path, streaming path, provider routing, or async cost/audit worker changes are planned.
- **Typed contracts**: PASS. Frontend TypeScript types will mirror the existing typed Go JSON response fields for usage and audit payloads.
- **Routing reliability**: N/A. The feature does not change model aliases, retries, fallbacks, rate limits, cooldowns, timeouts, or health-aware routing.
- **Multi-tenant spend governance**: PASS. The UI exposes API key-scoped usage/audit filters without changing enforcement semantics.
- **Observability**: PASS. Audit rows preserve request IDs, routing context, status, latency, tokens, cost, streaming flag, and redacted error context.
- **API capability UX parity**: PASS. The plan adds `ui/src/app/(dashboard)/usage` and navigation/dashboard entry points for existing usage/audit APIs.
- **Security and config**: PASS. Admin fetches use `LLMGOPHER_UI_ADMIN_API_KEY`; views must not render raw API keys, provider credentials, or request payloads.
- **Test and lint discipline**: PASS. UI tests cover fetch helpers, validation, filters, pages, and error states; Go tests/lint remain required if backend code changes.
- **Linter-first enforcement**: PASS. No new repeatable Go lint rule is needed; existing TypeScript/ESLint/Vitest checks are sufficient for this UI parity feature.

### Post-Design Gate

- **Upstream parity**: PASS. Contracts document existing admin endpoints and no backend behavior changes.
- **High-throughput runtime**: PASS. Queries remain operator-triggered admin reads; backend pagination/windowing limits remain intact.
- **Typed contracts**: PASS. `data-model.md` and `contracts/usage-audit-dashboard.md` define typed UI-facing shapes and validation rules.
- **Routing reliability**: N/A. Routing is read-only context in audit rows.
- **Multi-tenant spend governance**: PASS. Filters support API key, model, provider, status, and time range without exposing key secrets.
- **Observability**: PASS. UI uses existing audit context and documents unavailable/empty states.
- **API capability UX parity**: PASS. New admin UI surface closes the documented usage/audit gap.
- **Security and config**: PASS. Missing or unauthorized admin token is an explicit unavailable state.
- **Test and lint discipline**: PASS. Quickstart lists focused Vitest checks and full UI build/test verification.
- **Linter-first enforcement**: PASS. No additional deterministic lint enforcement is introduced.

## Project Structure

### Documentation (this feature)

```text
specs/32-ui-usage-audit-dashboard/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── usage-audit-dashboard.md
└── tasks.md              # Created later by /speckit-tasks
```

### Source Code (repository root)

```text
internal/
├── api/
│   ├── admin.go          # Existing GET /v1/admin/audit contract
│   └── admin_usage.go    # Existing GET /v1/admin/usage and /usage/daily contracts
└── storage/
    ├── audit_query.go    # Existing paginated audit query
    └── usage_store.go    # Existing grouped and daily usage queries

ui/
└── src/
    ├── app/(dashboard)/
    │   ├── page.tsx
    │   └── usage/
    │       ├── page.tsx
    │       └── page.test.tsx
    ├── components/
    │   └── usage/
    │       ├── AuditLogTable.tsx
    │       ├── UsageFilterForm.tsx
    │       ├── UsageSummaryTable.tsx
    │       └── UsageTrendTable.tsx
    ├── components/layout/
    │   └── sidebar-config.tsx
    └── lib/
        ├── analytics.ts
        ├── analytics.test.ts
        └── types.ts
```

**Structure Decision**: Use the existing Next.js dashboard layout and local component/test patterns. Keep backend Go files read-only unless implementation discovers a contract mismatch; the planned work is frontend parity over existing admin endpoints.

## Complexity Tracking

No constitution violations or additional complexity exceptions are required.
