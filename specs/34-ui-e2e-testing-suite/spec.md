# Feature Specification: UI End-to-End Testing Suite

**Feature Branch**: `[34-ui-e2e-testing-suite]`
**Created**: 2026-05-09
**Status**: Draft
**Input**: Operationalize `TESTING.md` — a multi-layer test suite (Playwright functional E2E + Applitools/Chromatic visual regression + axe-core accessibility) for the LLMGopher admin UI, run on every PR.

## Context

`TESTING.md` describes the desired end state. A first slice has already landed:

- Playwright is installed (`@playwright/test`, `@axe-core/playwright`) with `playwright.config.ts` driving two color-scheme projects (`light-comfy`, `dark-comfy`) against `npm run dev` on port 5173.
- Stable selectors: `data-testid="nav-{id}"` on every sidebar link (ids: `overview`, `logs`, `audit`, `providers`, `routes`, `guardrails`, `keys`, `teams`, `budgets`, `rate`, `settings`) and `data-testid="page-title"` on the topbar breadcrumb.
- `tests/e2e/navigation.spec.ts` covers shell navigation (3 passing tests, 1 fixme for command palette).

Everything else in `TESTING.md` is open. Subsequent stories should each land as an independently shippable slice.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Mock Backend for Deterministic E2E (Priority: P1)

The Playwright suite needs a deterministic backend. The UI uses Next.js **server actions** (see `ui/src/lib/actions.ts`) that `fetch` directly to `http://gateway:8080`, so MSW in the browser cannot intercept them — interception must be at the Node fetch layer or via a real mock HTTP server.

**Why this priority**: Every other test story depends on stable, seeded data. Without it, visual diffs flake on timestamps/spend numbers, and functional flows can't assert end states.

**Independent Test**: With the mock backend running, load `/providers`, `/keys`, `/logs`, `/budgets`, `/teams`, `/rate-limits`, `/audit` and confirm each renders deterministic seeded rows; reload and confirm identical bytes.

**Acceptance Scenarios**:

1. **Given** `LLMGOPHER_UI_BACKEND_MODE=mock`, **When** the Next.js app starts, **Then** server actions and any client fetches resolve against the mock instead of the real gateway.
2. **Given** the mock is seeded with a fixture set, **When** the same page is loaded twice, **Then** the rendered DOM (modulo intentionally-ignored regions) is byte-stable.
3. **Given** a test wants to simulate an error, **When** it tags a request with a header or query param, **Then** the mock returns the configured error envelope in OpenAI shape.

**Implementation hints** (informative, not prescriptive):
- Two viable approaches: (a) an in-process mock — a small Express/Hono server started by Playwright's `webServer` config that serves the admin endpoints in `ui/src/lib/actions.ts`; or (b) a `--mock` flag on the Go gateway binary that serves the same shapes from in-memory fixtures. Option (b) preserves contract fidelity at the cost of a Go build dependency in CI.
- Endpoints to mock at minimum (derived from `actions.ts`): `GET/POST/PATCH/DELETE /v1/admin/models`, `/v1/admin/providers`, `/v1/admin/keys`, `/v1/admin/keys/{id}/budget`, `/v1/admin/keys/{id}/budget/reset`, `/v1/admin/audit`, `/v1/admin/usage`, `/v1/admin/teams`, `/v1/admin/budgets`, `/v1/admin/rate-limits`, `/v1/admin/guardrails`, `/v1/admin/logs`.
- Fixture set must include: 2+ providers (one degraded), 5+ models, 3+ API keys (one near-cap, one over-cap), 20+ log rows mixing 2xx/4xx/5xx and one fallback row, 2 teams, 1 tripped rate-limit rule, 3 guardrails (mix of enabled/disabled), 10+ audit entries.
- Freeze "now" to a fixed ISO timestamp so relative-time formatters are deterministic.

---

### User Story 2 — Per-Surface Functional Coverage (Priority: P1)

Implement the functional groups laid out in `TESTING.md` §"Functional E2E Suite". Each group lands as its own spec file under `ui/tests/e2e/` and depends on Story 1.

**Why this priority**: Functional tests catch behavioral regressions (drawer open/close, form submit, optimistic rollback) that visual tests cannot.

**Independent Test**: Each surface's spec file can be run alone via `npx playwright test tests/e2e/<surface>.spec.ts` and passes against the mock backend.

**Acceptance Scenarios** (one per surface from `TESTING.md`):

1. **providers.spec.ts** — Add-OpenAI-provider wizard (3 steps), invalid base URL inline error with no POST sent, row appears after create. Requires `data-testid` on: `add-provider`, `provider-kind-{openai,anthropic,vertex,bedrock,cohere,generic}`, `wizard-next`, `wizard-create`, plus form labels `Display name`, `API key`. *Gap*: 3-step wizard does not yet exist — see Feature Gaps below.
2. **routes.spec.ts** — Switching a route to `fallback` renders a dashed secondary path; weight slider updates curve thickness on the diagram. Requires `data-testid` on: `route-{slug}`, plus the SVG strategy diagram. *Gap*: routes page is currently a stub.
3. **keys.spec.ts** — Rotate-key emits exactly one `POST` to `/admin/keys/{id}/rotate` and reveals the new prefix once. Hard-cap toggle flips a "near cap" pill to danger when budget tick crosses cap. Requires `data-testid` on: `rotate-key`, `confirm-rotate`, `one-time-key-reveal`, key-row identifiers.
4. **logs.spec.ts** — Filter `5xx` narrows table to error rows only; clicking a fallback row opens the inspector trace tab with the failed primary stage marked `data-failed="true"`. Requires `data-testid`: `filter-5xx`, `log-row`, `log-row-fallback`, `status`, `timeline-stage-primary`. *Gap*: logs page is a stub; request inspector drawer doesn't exist.
5. **budgets.spec.ts** — Team at 86% utilization shows warning indicator; key at hard cap returns `429` with `x-llmgopher-reason: budget_exceeded` (this last bit is a contract assertion, not UI — keep it co-located so cap behavior is verified end-to-end).
6. **guardrails.spec.ts** — Toggling `gr_jail` on persists across reload (`aria-checked="true"`). Requires `data-testid="toggle-{id}"`.
7. **rate-limits.spec.ts** — One rule renders a "tripped" pill when the mock reports it tripped.
8. **rbac.spec.ts** — `viewer` role hides destructive actions (`rotate-key`, `disable-key`). *Gap*: roles do not yet exist; see spec `24-rbac-jwt-auth/`.

**Selectors policy**: Prefer role/label queries; fall back to `data-testid` only where structure is ambiguous (table rows, modal triggers, drawer contents). Add testids in the same PR as the test that needs them — do not pre-seed the codebase.

---

### User Story 3 — Visual Regression Snapshots (Priority: P2)

Capture one snapshot per `screen × theme × density` combination per `TESTING.md` §"Visual Regression Suite". Pick **one** of: Applitools Eyes, Percy, or Chromatic. Recommendation: Applitools Playwright SDK because it slots into the existing `@playwright/test` runner without a second harness.

**Why this priority**: Visual diffs catch unintended pixel drift (theme regressions, density bugs) that functional tests miss; depends on Story 1 for determinism.

**Independent Test**: First run establishes baseline; second run on an unmodified branch produces zero diffs.

**Acceptance Scenarios**:

1. **Given** the snapshot matrix in `TESTING.md`, **When** the visual suite runs, **Then** every screen × theme × density tuple is captured.
2. **Given** sparklines, request IDs, timestamps, and the request flow strip exist on a page, **When** Applitools compares, **Then** those regions are excluded via `ignoreRegion`/`layoutRegion`.
3. **Given** CSS animations exist, **When** visual tests run, **Then** they are disabled via the inline `*, *::before, *::after { animation: none; transition: none; }` style block.
4. **Given** a real visual regression is introduced, **When** Applitools dashboard reviewer rejects it, **Then** the PR check fails.

**Density**: A "compact" density toggle does not yet exist in the UI. *Gap*: skip the `light-compact` project until added.

---

### User Story 4 — Accessibility Checks (Priority: P2)

Run `@axe-core/playwright` on every primary screen. Zero violations is the bar.

**Why this priority**: Catches regressions cheaply once the runner is up; strictly additive over functional coverage.

**Independent Test**: `tests/e2e/a11y.spec.ts` runs an `AxeBuilder` analysis on each route and asserts `violations` is empty.

**Acceptance Scenarios**:

1. **Given** the overview, providers, routes, keys, teams, budgets, rate-limits, logs, guardrails, audit, settings pages, **When** axe runs, **Then** `results.violations` is `[]` for each.
2. **Given** the active sidebar item, **When** axe and DOM assertions run, **Then** it has `aria-current="page"` (already implemented; assert it).
3. **Given** any drawer is open, **When** Tab is pressed past the last focusable element, **Then** focus wraps inside the drawer (focus trap).
4. **Given** any sortable table header, **When** focused and activated, **Then** the sort state is announced via `aria-sort`.

---

### User Story 5 — CI Integration (Priority: P3)

Wire the new suite into GitHub Actions. The repo's existing CI does not run Playwright today.

**Why this priority**: Local-only tests rot. Block PR merge on functional failure, rejected visual diff, or new a11y violation.

**Independent Test**: A PR that intentionally breaks the navigation test fails the CI check `ui-e2e`.

**Acceptance Scenarios**:

1. **Given** a PR touches `ui/**`, **When** CI runs, **Then** `npm ci`, `npm run build`, and `npx playwright test` execute against the mock backend.
2. **Given** Applitools is configured, **When** CI runs, **Then** `APPLITOOLS_API_KEY` and `APPLITOOLS_BATCH_ID=${{ github.run_id }}` are set.
3. **Given** Playwright traces exist on retry, **When** a test fails, **Then** the trace is uploaded as a workflow artifact for triage.
4. **Given** the suite passes, **When** the PR is reviewed, **Then** the merge is unblocked.

---

## Feature Gaps Surfaced by `TESTING.md`

These are UI capabilities `TESTING.md` assumes exist but currently do not. Placeholder pages with `Coming soon.` copy are now specified in `specs/35-ui-coming-soon-surfaces/`:

- **Command palette** (⌘K) — referenced in `TESTING.md` §1; topbar shows a `⌘K` kbd hint but no palette is wired up. Test left as `test.fixme` in `navigation.spec.ts`.
- **Add-Provider wizard** (3-step) — `TESTING.md` §2 assumes a wizard with kind picker, credential validation, and model picks. Current `ui/src/components/CreateProviderModal.tsx` (if present) is a flat form.
- **Routes page** with strategy switcher and SVG diagram — see `specs/35-ui-coming-soon-surfaces/surface-specs/routes.md`.
- **Logs page** with status filters and request inspector drawer (trace/prompt/response/headers tabs) — see `specs/35-ui-coming-soon-surfaces/surface-specs/logs.md`.
- **Teams grid** and **per-team budget alerts** with 85% threshold indicator — see `specs/35-ui-coming-soon-surfaces/surface-specs/teams.md`.
- **Budgets page** with team/key spend policy controls and warning states — see `specs/35-ui-coming-soon-surfaces/surface-specs/budgets.md`.
- **Guardrails toggles persistence** — see `specs/35-ui-coming-soon-surfaces/surface-specs/guardrails.md`.
- **Audit page** with filterable request/admin history — see `specs/35-ui-coming-soon-surfaces/surface-specs/audit.md`.
- **Rate-limits "tripped" pill** — see `specs/35-ui-coming-soon-surfaces/surface-specs/rate-limits.md`.
- **Settings page** with Gateway Profile, Security, Notifications, and Display cards — see `specs/35-ui-coming-soon-surfaces/surface-specs/settings.md`.
- **RBAC viewer role** — depends on `24-rbac-jwt-auth`.
- **Density toggle** (`compact`) — referenced by `TESTING.md` Playwright project `light-compact`.

## Non-Goals

- Replacing Vitest unit tests under `ui/src/**/*.test.ts(x)`. Those remain the unit layer.
- Hitting a real upstream provider (OpenAI/Anthropic/Vertex). The mock backend serves canned responses only.
- Performance/load testing.

## Open Questions

1. **Visual regression vendor**: Applitools, Percy, or Chromatic? Cost and dashboard-review workflow differ. Recommend Applitools per `TESTING.md`; needs sign-off + secret provisioning before Story 3.
2. **Mock backend approach**: in-process Node mock (faster, drifts from real gateway) vs. `gateway --mock` flag (slower CI, contract-true). Recommend starting with in-process and migrating if drift bites.
3. **CI runner**: macOS runners are slow and expensive on GH Actions. Default to `ubuntu-latest` for E2E unless a Mac-only bug surfaces.
