# Implementation Plan: UI Coming Soon Surfaces

**Branch**: `main` (planning artifact for `35-ui-coming-soon-surfaces`) | **Date**: 2026-05-09 | **Spec**: [`spec.md`](./spec.md)
**Input**: Feature specification from `specs/35-ui-coming-soon-surfaces/spec.md`

## Summary

Replace the eight admin UI placeholder pages (`/logs`, `/audit`, `/routes`, `/guardrails`, `/teams`, `/budgets`, `/rate-limits`, and `/settings`) with scoped, typed, testable surfaces. The first implementation is a Next.js admin UI change that consumes existing mock contracts where available, defines missing route/settings contracts before writes, encodes Logs/Audit filters in URL query parameters, and keeps production mutation controls unavailable until the corresponding gateway admin APIs are reconciled.

## Technical Context

**Language/Version**: TypeScript 5.x in `ui/`; no Go runtime changes planned for this feature.  
**Primary Dependencies**: Next.js 15.5, React 19.2, existing UI component primitives, Playwright, `@axe-core/playwright`, Applitools Eyes Playwright SDK.  
**Storage**: N/A for production UI state beyond URL query parameters and local-only display preferences; E2E mock state remains in `ui/tests/mock/state.ts`.  
**Testing**: Vitest/React Testing Library for component logic, Playwright for functional E2E, `@axe-core/playwright` for accessibility, visual snapshots through the existing E2E visual layer.  
**Target Platform**: Admin UI in modern browsers, built by Next.js and exercised headlessly on Linux CI.  
**Project Type**: Web application under `ui/`, backed by admin HTTP contracts and deterministic mock fixtures.  
**Performance Goals**: Initial page render and filter interactions should remain client-side responsive for the seeded E2E datasets; no new gateway request hot-path work, provider calls, or blocking audit/cost behavior.  
**Constraints**: Production mutation controls stay read-only/unavailable until real admin contracts exist; credentials and secret-like payloads must be redacted before rendering; Logs/Audit filters must survive reload/share via query parameters; E2E selectors must remain minimal and accessibility-first.  
**Scale/Scope**: Eight replacement pages, eight page-level surface specs, fixture-backed functional/a11y/visual coverage for the page states named in `TESTING.md`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity** — PASS. The feature adds admin UI surfaces over existing LLMGopher capabilities and mock contracts; it does not change OpenAI-compatible client APIs. Route, budget, guardrail, rate-limit, team, audit, and logs behavior stays aligned with the existing backend feature specs and real contracts when present.
- **High-throughput runtime** — PASS. No gateway hot-path, provider, middleware, streaming, cache, or async cost/audit worker changes are planned. UI pages call admin/mock endpoints only.
- **Typed contracts** — PASS. Each rendered payload is represented by a TypeScript UI contract or an existing `ui/tests/mock/types.ts`/fixture type before use; dynamic raw JSON display is out of scope.
- **Routing reliability** — PASS. The Routes page is an operator surface for fallback/weighted/latency/single-provider policy inspection. Runtime routing behavior remains unchanged until a reconciled admin mutation contract exists.
- **Multi-tenant spend governance** — PASS. Teams, Budgets, and Rate Limits expose key/team/model scope context and warning/tripped states without changing enforcement semantics.
- **Observability** — PASS. Logs and Audit improve request/audit observability while preserving redaction and request/provider outcome context.
- **API capability UX parity** — PASS. This feature exists to close UI gaps for already-specified gateway capabilities; unavailable API-backed writes are documented with follow-up triggers.
- **Security and config** — PASS. Secrets, authorization headers, provider credentials, raw API keys, prompts, responses, and secret-like metadata are redacted or truncated before display. No runtime config precedence changes.
- **Test and lint discipline** — PASS. UI changes use `npm run lint`, Vitest where component logic is introduced, and focused Playwright coverage. `golangci-lint` is N/A because no Go code is planned.
- **Linter-first enforcement** — PASS. No new repeatable Go rule is introduced. UI selector discipline remains covered by existing Playwright/ESLint conventions from the E2E plan.

Post-design check: all gates remain PASS or N/A after Phase 1 artifacts. No Complexity Tracking entries are required.

## Project Structure

### Documentation (this feature)

```text
specs/35-ui-coming-soon-surfaces/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── admin-audit.md
│   ├── admin-budgets.md
│   ├── admin-guardrails.md
│   ├── admin-logs.md
│   ├── admin-rate-limits.md
│   ├── admin-routes.md
│   ├── admin-settings.md
│   └── admin-teams.md
├── surface-specs/
│   ├── audit.md
│   ├── budgets.md
│   ├── guardrails.md
│   ├── logs.md
│   ├── rate-limits.md
│   ├── routes.md
│   ├── settings.md
│   └── teams.md
└── tasks.md             # Phase 2 output from /speckit-tasks
```

### Source Code (repository root)

```text
ui/
├── src/app/(dashboard)/
│   ├── audit/page.tsx
│   ├── budgets/page.tsx
│   ├── guardrails/page.tsx
│   ├── logs/page.tsx
│   ├── rate-limits/page.tsx
│   ├── routes/page.tsx
│   ├── settings/page.tsx
│   └── teams/page.tsx
├── src/components/
│   ├── audit/
│   ├── budgets/
│   ├── guardrails/
│   ├── logs/
│   ├── rate-limits/
│   ├── routes/
│   ├── settings/
│   └── teams/
├── src/lib/
│   ├── admin-surface-contracts.ts
│   ├── redaction.ts
│   └── query-state.ts
└── tests/
    ├── e2e/
    │   ├── audit.spec.ts
    │   ├── budgets.spec.ts
    │   ├── guardrails.spec.ts
    │   ├── logs.spec.ts
    │   ├── rate-limits.spec.ts
    │   ├── routes.spec.ts
    │   ├── settings.spec.ts
    │   └── teams.spec.ts
    ├── fixtures/
    └── mock/handlers/
```

**Structure Decision**: Keep all implementation in `ui/`, using one surface-oriented component folder per placeholder page plus shared redaction/query-state/contract helpers in `ui/src/lib/`. Existing E2E fixtures and mock handlers remain the deterministic source for the first implementation.

## Complexity Tracking

> No constitution violations. Section intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
