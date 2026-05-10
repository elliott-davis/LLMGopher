# Implementation Plan: UI End-to-End Testing Suite

**Branch**: `033-ui-model-rate-limits` (working branch; this feature can land on its own branch later) | **Date**: 2026-05-09 | **Spec**: [`spec.md`](./spec.md)
**Input**: Feature specification from `specs/34-ui-e2e-testing-suite/spec.md`

## Summary

Operationalize `TESTING.md`: a Playwright-driven functional E2E suite, an Applitools-backed visual regression layer, and an `@axe-core/playwright` accessibility check, all run on every PR via GitHub Actions. The first slice (Playwright scaffolding, color-scheme projects, sidebar `data-testid`s, and a passing navigation spec) has already shipped. The remaining work splits along five user stories: a deterministic mock backend (P1), per-surface functional coverage (P1), visual regression (P2), accessibility (P2), and CI integration (P3).

## Technical Context

**Language/Version**: TypeScript 5.x (UI). No Go changes unless Story 1 chooses the `gateway --mock` mock approach.
**Primary Dependencies**: Next.js 15.5, React 19.2, `@playwright/test` ^1.48, `@axe-core/playwright` ^4.10, Applitools Eyes Playwright SDK (TBD per Story 3).
**Storage**: N/A — fixtures live in test code; mock backend keeps state in-process per worker.
**Testing**: Playwright (E2E + a11y), Vitest (unit, unchanged). Vitest excludes `tests/e2e/**`.
**Target Platform**: Headless Chromium on `ubuntu-latest` GitHub Actions runners; local dev on macOS/Linux.
**Project Type**: Web application — UI subdirectory `ui/` only; no backend code paths exercised in this feature except the optional `gateway --mock` flag.
**Performance Goals**: Full suite (~80 functional + ~30 visual + ~12 a11y assertions) completes in under 6 minutes on CI. Per-test timeout 30s.
**Constraints**: Suite MUST be deterministic — no flaky time-based assertions, no live provider calls, no network egress beyond `localhost`. CSS animations disabled during visual runs.
**Scale/Scope**: 11 admin UI surfaces (overview, logs, audit, providers, routes, guardrails, keys, teams, budgets, rate-limits, settings) × 2 themes = 22 baseline visual snapshots, plus drawer/wizard variants per `TESTING.md` matrix.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Upstream parity** — N/A. Feature is testing infrastructure; no API surface changes. Mock backend (Story 1) MUST mirror existing LLMGopher admin endpoint shapes — that is the only parity concern, captured in `contracts/`.
- **High-throughput runtime** — N/A. No hot-path code touched.
- **Typed contracts** — PASS. Fixtures and mock handlers will reuse the existing TypeScript types in `ui/src/lib/types.ts` rather than redefining shapes.
- **Routing reliability** — N/A.
- **Multi-tenant spend governance** — N/A for this feature. Test fixtures will exercise budget/rate-limit/RBAC UX once those features ship (cross-references in `spec.md` "Feature Gaps").
- **Observability** — N/A. Tests assert on observable UI state; no logging/metrics changes.
- **API capability UX parity** — PASS. This feature *is* a UX-validation mechanism; it strengthens parity by catching UI regressions when API capabilities change. No new API capability ships here, so no new UI surface is owed.
- **Security and config** — PASS. Applitools API key is a CI secret; mock backend ships no real provider credentials. Fixture data MUST NOT include any real-looking provider keys (use clearly-fake `sk-test-…` strings).
- **Test and lint discipline** — PASS. Feature *adds* test coverage. Vitest unit tests remain. ESLint covers TS in `ui/`. No Go changes, so `golangci-lint` is N/A here.
- **Linter-first enforcement** — PASS. Selectors stability is enforced by an ESLint rule recommendation (see research): forbid query-by-text in E2E specs, prefer `getByRole`/`getByTestId`.

All gates PASS or N/A. No Complexity Tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/34-ui-e2e-testing-suite/
├── plan.md              # This file
├── research.md          # Phase 0 — vendor + mock approach + selector policy decisions
├── data-model.md        # Phase 1 — fixture entities for the mock backend
├── quickstart.md        # Phase 1 — "run the tests in 60 seconds" guide
├── contracts/           # Phase 1 — mock backend endpoint contracts (one file per group)
│   ├── admin-providers.md
│   ├── admin-models.md
│   ├── admin-keys.md
│   ├── admin-budgets.md
│   ├── admin-teams.md
│   ├── admin-rate-limits.md
│   ├── admin-guardrails.md
│   ├── admin-logs.md
│   └── admin-audit.md
└── tasks.md             # Phase 2 (created later by /speckit-tasks)
```

### Source Code (repository root)

Only the `ui/` subtree is touched.

```text
ui/
├── playwright.config.ts                 # ✅ landed — light/dark projects, webServer
├── package.json                         # ✅ landed — test:e2e script, deps
├── vitest.config.ts                     # ✅ landed — excludes tests/e2e
├── tests/
│   ├── e2e/
│   │   ├── navigation.spec.ts           # ✅ landed — shell + nav coverage
│   │   ├── providers.spec.ts            # Story 2
│   │   ├── routes.spec.ts               # Story 2 (depends on routes-page feature)
│   │   ├── keys.spec.ts                 # Story 2
│   │   ├── logs.spec.ts                 # Story 2 (depends on logs-page feature)
│   │   ├── budgets.spec.ts              # Story 2
│   │   ├── teams.spec.ts                # Story 2
│   │   ├── rate-limits.spec.ts          # Story 2
│   │   ├── guardrails.spec.ts           # Story 2
│   │   ├── audit.spec.ts                # Story 2
│   │   ├── settings.spec.ts             # Story 2
│   │   ├── a11y.spec.ts                 # Story 4
│   │   └── visual.spec.ts               # Story 3
│   ├── fixtures/
│   │   ├── seed.ts                      # Default deterministic seed
│   │   ├── providers.ts
│   │   ├── keys.ts
│   │   ├── logs.ts
│   │   └── ...                          # one per entity (see data-model.md)
│   └── mock/
│       ├── server.ts                    # Story 1 — Hono/Express mock entry
│       ├── handlers/                    # one file per endpoint group (mirrors contracts/)
│       └── state.ts                     # in-memory store, reset between tests
└── src/components/layout/
    ├── Sidebar.tsx                      # ✅ data-testid wired
    ├── Topbar.tsx                       # ✅ data-testid wired
    └── sidebar-config.tsx               # ✅ testId field added

.github/workflows/
└── ui-e2e.yml                           # Story 5
```

**Structure Decision**: Single `ui/tests/e2e/` directory keyed by surface, with mock backend + fixtures siblings under `ui/tests/`. This matches the existing convention (Vitest tests live next to source, Playwright tests in a separate root) and keeps the mock backend co-located with the only code that consumes it.

## Complexity Tracking

> No constitution violations. Section intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
