---
description: "Task list for UI End-to-End Testing Suite (spec 34)"
---

# Tasks: UI End-to-End Testing Suite

**Input**: Design documents from `/specs/34-ui-e2e-testing-suite/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: This feature is *itself* a testing infrastructure project — every implementation task IS a test or test-supporting artifact. There is no parallel "tests for the tests" layer.

**Organization**: Tasks are grouped by user story (US1–US5) so each can be picked up and shipped independently. Setup and Foundational phases land first.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different file, no dependency on incomplete tasks — safe to run in parallel.
- **[Story]**: US1–US5 from spec.md.

## Path Conventions

All paths are under `ui/` unless noted. The repo root is `/Users/elliottdavis/src/LLMGopher`. Mock backend lives at `ui/tests/mock/`, fixtures at `ui/tests/fixtures/`, specs at `ui/tests/e2e/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Land the Playwright runner, color-scheme projects, sidebar selectors, and a passing navigation spec. **Mostly already shipped** — remaining tasks finish the scaffolding.

- [x] T001 Install `@playwright/test` ^1.48 and `@axe-core/playwright` ^4.10 in `ui/package.json` and add `test:e2e` / `test:e2e:ui` scripts.
- [x] T002 Add `ui/playwright.config.ts` with `light-comfy` and `dark-comfy` Chromium projects, `webServer` boot of `next dev` on port 5173.
- [x] T003 Add `testId: string` field to `NavItem` in `ui/src/components/layout/sidebar-config.tsx` and populate every nav entry.
- [x] T004 Wire `data-testid="nav-{testId}"` onto `<Link>` in `ui/src/components/layout/Sidebar.tsx` and `data-testid="page-title"` onto the breadcrumb in `ui/src/components/layout/Topbar.tsx`.
- [x] T005 Exclude `tests/e2e/**` from Vitest in `ui/vitest.config.ts`.
- [x] T006 Land first navigation spec in `ui/tests/e2e/navigation.spec.ts` covering sidebar visibility, click-through, `aria-current`, and a `test.fixme` for command palette.
- [x] T007 Add an `npm run test:e2e:install` convenience script in `ui/package.json` that runs `playwright install chromium` (used by CI cache-miss path).
- [x] T008 [P] Add `eslint-plugin-playwright` (or equivalent) and an ESLint config block forbidding raw CSS selectors in `ui/tests/e2e/**` per research.md Decision 3.

**Checkpoint**: Playwright runs; `npx playwright test --project=light-comfy` passes 3 / fixmes 1.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Refactors and shared utilities that every story below depends on.

**⚠️ CRITICAL**: Stories US1–US5 cannot ship without these.

- [x] T009 Refactor `ui/src/lib/actions.ts` to read `process.env.LLMGOPHER_GATEWAY_BASE ?? "http://gateway:8080"` instead of hardcoding the URL constant. Apply to every endpoint constant in the file.
- [x] T010 [P] Audit `ui/src/lib/**` for any other hardcoded `http://gateway:8080` references and migrate them through the same env helper.
- [x] T011 [P] Add `ui/tests/support/test-utils.ts` exporting Playwright fixture helpers: `freezeTime(page)` (patches `Date.now` per research Decision 6) and `disableAnimations(page)` (injects the `*, *::before, *::after { animation: none; transition: none; }` style).
- [x] T012 [P] Add `ui/tests/support/mock-port.ts` exporting `MOCK_PORT = 8787` and a helper for the env block passed to `webServer` (reused by Story 1 to set `LLMGOPHER_GATEWAY_BASE=http://127.0.0.1:8787`).

**Checkpoint**: Server actions are env-driven; shared Playwright helpers exist.

---

## Phase 3: User Story 1 — Mock Backend for Deterministic E2E (Priority: P1) 🎯 MVP

**Goal**: Boot a deterministic Hono mock backend alongside `next dev` so every subsequent test sees the same data.

**Independent Test**: With the mock running, `GET http://127.0.0.1:8787/v1/admin/providers` returns the fixture seed; loading `/providers` in a browser renders the seeded rows; reload produces byte-identical DOM (modulo ignored regions).

### Implementation for User Story 1

- [x] T013 [P] [US1] Add `hono` (or `express` — Hono recommended for size) to `ui/package.json` devDependencies.
- [x] T014 [P] [US1] Define mock-backend types re-exported from `ui/src/lib/types.ts` in `ui/tests/mock/types.ts` (no new shapes — only re-exports plus the `LogRow`, `RateLimitRule`, `Guardrail`, `Team` types listed in `data-model.md` that don't yet exist in `src/lib/types.ts`).
- [x] T015 [P] [US1] Create `ui/tests/fixtures/providers.ts` per `data-model.md` (3 providers: healthy/degraded/offline).
- [x] T016 [P] [US1] Create `ui/tests/fixtures/models.ts` (6 models, one with `rate_limit`).
- [x] T017 [P] [US1] Create `ui/tests/fixtures/keys.ts` (4 keys including `key_near_cap` at 86% and `key_over_cap` at 100%).
- [x] T018 [P] [US1] Create `ui/tests/fixtures/teams.ts` (`team_research` at 0.86 utilization, `team_platform` at 0.40).
- [x] T019 [P] [US1] Create `ui/tests/fixtures/budgets.ts` (per-team and per-key budget rows derived from keys/teams).
- [x] T020 [P] [US1] Create `ui/tests/fixtures/rate-limits.ts` (3 rules, exactly one with `tripped: true`).
- [x] T021 [P] [US1] Create `ui/tests/fixtures/guardrails.ts` (`gr_jail`, `gr_pii`, `gr_secrets` with mixed `enabled`).
- [x] T022 [P] [US1] Create `ui/tests/fixtures/logs.ts` (20 rows, mixed status, one row id `log_fallback` with multi-stage `provider_chain`).
- [x] T023 [P] [US1] Create `ui/tests/fixtures/audit.ts` (10 entries).
- [x] T024 [US1] Create `ui/tests/fixtures/seed.ts` composing the entity fixtures into a single `seed()` snapshot returning `structuredClone`-able state.
- [x] T025 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-providers.ts` per `contracts/admin-providers.md` (GET, POST, PATCH, DELETE).
- [x] T026 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-models.ts` per `contracts/admin-models.md`.
- [x] T027 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-keys.ts` per `contracts/admin-keys.md`, including the rotate endpoint and one-time-reveal semantics.
- [x] T028 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-budgets.ts` per `contracts/admin-budgets.md`.
- [x] T029 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-teams.ts` per `contracts/admin-teams.md`.
- [x] T030 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-rate-limits.ts` per `contracts/admin-rate-limits.md`.
- [x] T031 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-guardrails.ts` per `contracts/admin-guardrails.md` with reload-persistence inside a worker.
- [x] T032 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-logs.ts` per `contracts/admin-logs.md`, including the `?status=` filter and the `/{id}` detail with redacted headers.
- [x] T033 [P] [US1] Implement handlers in `ui/tests/mock/handlers/admin-audit.ts` per `contracts/admin-audit.md`.
- [x] T034 [P] [US1] Implement the chat-completions cap-enforcement contract in `ui/tests/mock/handlers/chat-completions.ts` returning `429 budget_exceeded` per `contracts/admin-keys.md`. All other chat-completions traffic returns `501`.
- [x] T035 [US1] Create `ui/tests/mock/state.ts` holding the in-memory store and a `reset()` function that re-applies the seed snapshot via `structuredClone`.
- [x] T036 [US1] Create `ui/tests/mock/server.ts` that mounts every handler on Hono, listens on `MOCK_PORT`, and supports the `x-mock-error` and `?__mock_latency_ms` knobs from `contracts/README.md`. Add a CLI entry so `node tests/mock/server.ts` works for manual testing.
- [x] T037 [US1] Update `ui/playwright.config.ts` `webServer` to a two-process boot: the mock server first (with `url: http://127.0.0.1:8787/healthz`), then `next dev` with `LLMGOPHER_GATEWAY_BASE=http://127.0.0.1:8787` and `LLMGOPHER_UI_ADMIN_API_KEY=test-token`.
- [x] T038 [US1] Add `ui/tests/e2e/mock-smoke.spec.ts` that hits `/admin/providers`, `/admin/keys`, `/admin/logs`, `/admin/teams`, `/admin/budgets`, `/admin/rate-limits`, `/admin/guardrails`, `/admin/audit` from the page context and asserts each renders the seeded fixture row count. This is the independent-test gate for US1.

**Checkpoint**: `npx playwright test tests/e2e/mock-smoke.spec.ts` passes. The dev server hits the mock backend, every page loads, every fixture surface is reachable. **MVP for the rest of the suite.**

---

## Phase 4: User Story 2 — Per-Surface Functional Coverage (Priority: P1)

**Goal**: One spec file per admin surface, covering the flows in `TESTING.md` §"Functional E2E Suite".

**Independent Test**: Each spec file passes alone via `npx playwright test tests/e2e/<surface>.spec.ts` against the US1 mock.

> Tasks marked **[blocked]** depend on a UI feature that doesn't yet exist. Land the spec under `test.fixme` so it documents the intent and unblocks the moment the feature ships.

### Surface specs

- [x] T039 [US2] Add `data-testid` to `ui/src/components/CreateProviderModal.tsx` (or its successor) for: `add-provider`, `provider-kind-{openai,anthropic,vertex,bedrock,cohere,generic}`, `wizard-next`, `wizard-create`. (Touches one file; sequence before T040.)
- [x] T040 [US2] `ui/tests/e2e/providers.spec.ts` — add-provider happy path (3-step wizard from `TESTING.md` §2), invalid base URL inline error (no POST sent), new row visible. Mark steps 2/3 `test.fixme` if the wizard is still flat (cross-ref Feature Gap "Add-Provider wizard").
- [x] T041 [P] [US2] `ui/tests/e2e/routes.spec.ts` — strategy switcher renders dashed secondary path, weight slider updates curve thickness. Mark file `test.fixme` until routes page ships (cross-ref Feature Gap).
- [x] T042 [US2] Add `data-testid` to `ui/src/components/ProviderRowActions.tsx` and the keys table for: `rotate-key`, `confirm-rotate`, `one-time-key-reveal`, plus row-level identifiers `key-row-{id}`. (Sequence before T043.)
- [x] T043 [US2] `ui/tests/e2e/keys.spec.ts` — rotate emits exactly one POST to `/admin/keys/{id}/rotate`, one-time reveal shown once, hard-cap toggle flips warn-pill to danger when budget tick crosses cap.
- [x] T044 [P] [US2] `ui/tests/e2e/logs.spec.ts` — `5xx` filter narrows to error rows; clicking `log_fallback` opens inspector with `[data-testid="timeline-stage-primary"]` exposing `data-failed="true"`. `test.fixme` until logs surface ships.
- [x] T045 [P] [US2] `ui/tests/e2e/budgets.spec.ts` — `team_research` shows `data-testid="team-research-warn"`; cap-exceeded contract: `POST /v1/chat/completions` with `key_over_cap` returns `429` with `x-llmgopher-reason: budget_exceeded`.
- [x] T046 [P] [US2] `ui/tests/e2e/teams.spec.ts` — teams grid populated, both seeded teams render.
- [x] T047 [P] [US2] `ui/tests/e2e/rate-limits.spec.ts` — rule with `tripped: true` shows the "tripped" pill. `test.fixme` if the surface from spec 33 has not landed.
- [x] T048 [P] [US2] `ui/tests/e2e/guardrails.spec.ts` — toggling `gr_jail` to on persists across reload (`aria-checked="true"`).
- [x] T049 [P] [US2] `ui/tests/e2e/audit.spec.ts` — audit list populated; date filter narrows results.
- [x] T050 [P] [US2] `ui/tests/e2e/settings.spec.ts` — all four settings cards render.
- [x] T051 [P] [US2] `ui/tests/e2e/rbac.spec.ts` — viewer role hides `rotate-key` and `disable-key`. Whole file `test.fixme` until spec `24-rbac-jwt-auth` ships.

**Checkpoint**: Every admin surface has an E2E spec. Implemented surfaces pass green; pending surfaces document expected behavior via `test.fixme`.

---

## Phase 5: User Story 3 — Visual Regression Snapshots (Priority: P2)

**Goal**: One Applitools snapshot per `screen × theme` per `TESTING.md` matrix, with deterministic suppressions.

**Independent Test**: First run establishes baseline; second run on an unmodified branch produces zero diffs.

- [x] T052 [US3] Add `@applitools/eyes-playwright` to `ui/package.json` devDependencies and run `npx eyes-setup` once to drop the integration helpers.
- [x] T053 [US3] Add `APPLITOOLS_API_KEY` to the project's GitHub Actions secrets and document in `ui/tests/e2e/visual.spec.ts` header comment.
- [x] T054 [US3] Create `ui/tests/e2e/visual.spec.ts` covering the snapshot matrix from `TESTING.md` §"Snapshots to capture": Overview (cards/charts/tables), Providers (populated, degraded, offline), Add-Provider drawer (3 steps), Routes (each strategy), API Keys (populated, drawer tabs), Logs (mixed/fallback/empty), Request inspector (trace success/fallback, prompt, response, headers), Teams, Budgets (near cap, over cap), Rate limits (tripped), Guardrails (mix), Audit, Settings.
- [x] T055 [P] [US3] Apply the `ignoreRegion` / `layoutRegion` suppression list from `TESTING.md` §"Suppressing flakiness" to the visual spec: `.spark`, `[data-flow-strip]`, `.id`, `table.tbl tbody`. Add new suppressions only when justified.
- [x] T056 [P] [US3] Apply `disableAnimations(page)` (T011) to every visual test before `Eyes.check`.
- [x] T057 [US3] Document the baseline-approval workflow (Applitools dashboard, who reviews) in `ui/tests/README.md`.

**Checkpoint**: Re-running visual on a clean branch reports zero diffs.

---

## Phase 6: User Story 4 — Accessibility Checks (Priority: P2)

**Goal**: Zero axe violations on every primary route + drawer-focus-trap and aria-sort assertions.

**Independent Test**: `npx playwright test tests/e2e/a11y.spec.ts` passes on every project.

- [x] T058 [US4] Create `ui/tests/e2e/a11y.spec.ts` running `AxeBuilder({ page }).analyze()` on `/`, `/logs`, `/audit`, `/providers`, `/routes`, `/guardrails`, `/keys`, `/teams`, `/budgets`, `/rate-limits`, `/settings`. Assert `violations` is empty for each.
- [x] T059 [P] [US4] Add a focus-trap test to `ui/tests/e2e/a11y.spec.ts` (or split into `a11y-focus.spec.ts`) for any open drawer: Tab past the last focusable wraps inside the drawer; closing returns focus to the trigger.
- [x] T060 [P] [US4] Add an aria-sort test asserting that activating a sortable table header announces the new sort state via `aria-sort`.
- [x] T061 [US4] Document any rule the dark theme legitimately fails (per research Decision 8) in `ui/tests/README.md` and gate it behind a tracked design-debt issue link rather than a silent disable.

**Checkpoint**: `a11y.spec.ts` passes on both `light-comfy` and `dark-comfy` projects.

---

## Phase 7: User Story 5 — CI Integration (Priority: P3)

**Goal**: Run the full suite on every PR that touches `ui/**`; block merge on failure.

**Independent Test**: A PR that intentionally breaks `navigation.spec.ts` fails the `ui-e2e` check.

- [x] T062 [US5] Create `.github/workflows/ui-e2e.yml`: trigger on `pull_request` paths `ui/**`; runs `npm ci`, `npm run build`, `npx playwright install --with-deps chromium`, `npx playwright test`. Cache `~/.npm` and `~/.cache/ms-playwright` keyed by `ui/package-lock.json`.
- [x] T063 [P] [US5] Wire `APPLITOOLS_API_KEY: ${{ secrets.APPLITOOLS_API_KEY }}` and `APPLITOOLS_BATCH_ID: ${{ github.run_id }}` into the workflow env.
- [x] T064 [P] [US5] Upload Playwright trace + report artifacts on failure (`actions/upload-artifact@v4`, paths `ui/playwright-report` and `ui/test-results`).
- [ ] T065 [US5] Configure GitHub branch protection for `main` to require the `ui-e2e` check (manual step — document in PR description).

**Checkpoint**: A failing PR cannot merge.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [x] T066 [P] Update `TESTING.md` to point to `specs/34-ui-e2e-testing-suite/quickstart.md` for the canonical run instructions; keep `TESTING.md` as the high-level intent doc.
- [x] T067 [P] Add `ui/tests/README.md` summarizing the directory layout (e2e/, fixtures/, mock/, support/) and the conventions from research.md.
- [x] T068 [P] Run `npx playwright test` end-to-end on a clean checkout; record baseline timing in the PR description.
- [ ] T069 [P] After each cross-referenced feature ships (command palette, add-provider wizard, routes, logs, RBAC, density toggle), remove the corresponding `test.fixme` and verify the now-active test passes.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)** — mostly done; T007–T008 can land anytime.
- **Phase 2 (Foundational)** — T009 unblocks every story below; T010–T012 unblock specific stories. Must complete before Phase 3 onward.
- **Phase 3 (US1, P1) — MVP** — depends on Phase 2. Blocks every other functional/visual story.
- **Phase 4 (US2, P1)** — depends on Phase 3.
- **Phase 5 (US3, P2)** — depends on Phase 3 (needs deterministic data). Independent of Phase 4.
- **Phase 6 (US4, P2)** — depends on Phase 3. Independent of Phases 4 and 5.
- **Phase 7 (US5, P3)** — depends on Phases 3+4+6 producing a stable green run; can land alongside Phase 5 since visual diffs are reviewer-gated.
- **Phase 8 (Polish)** — last; depends on the stories you actually shipped.

### Within Each Story

- Fixtures (T015–T023) before the seed composer (T024) before handlers (T025–T034) before the server (T035–T036) before the smoke spec (T038).
- `data-testid` additions (T039, T042) before the spec that consumes them (T040, T043).

### Parallel Opportunities

- T015–T023 (fixtures) in parallel.
- T025–T034 (handlers) in parallel.
- T041, T044, T046–T050 (per-surface specs) in parallel after Phase 3 lands.
- Phases 5 and 6 can run in parallel after Phase 3.

---

## Parallel Example: User Story 1 fixture set

```bash
# After T013–T014 land, the 9 fixture files have no inter-dependencies:
Task: "Create ui/tests/fixtures/providers.ts"
Task: "Create ui/tests/fixtures/models.ts"
Task: "Create ui/tests/fixtures/keys.ts"
Task: "Create ui/tests/fixtures/teams.ts"
Task: "Create ui/tests/fixtures/budgets.ts"
Task: "Create ui/tests/fixtures/rate-limits.ts"
Task: "Create ui/tests/fixtures/guardrails.ts"
Task: "Create ui/tests/fixtures/logs.ts"
Task: "Create ui/tests/fixtures/audit.ts"
```

---

## Implementation Strategy

### MVP First (Setup + Foundational + User Story 1)

1. Finish Phase 1 (T007, T008).
2. Land Phase 2 (T009 is the blocker; T010–T012 cleanup).
3. Land Phase 3 (US1 mock backend) — this is the MVP. Stop and validate via `mock-smoke.spec.ts`.
4. **Decision point**: ship US1 alone, then incrementally add US2/3/4/5.

### Incremental Delivery

- After US1 lands, every additional spec is purely additive — no story breaks the others.
- Visual (US3) and a11y (US4) are independent of each other and can be tackled by different developers in parallel.
- CI (US5) lands last so the green-bar is a real signal, not an early aspiration.

### Surface-by-surface within US2

When picking up US2 piecemeal, prioritize surfaces whose UI features already exist (providers, keys, guardrails, budgets, audit, settings). Park `test.fixme` files for routes, logs, RBAC, and the wizard until those features ship. Each removal of a `test.fixme` is its own small PR.

---

## Notes

- [P] tasks = different files, no dependencies.
- Every spec file uses the role/label-first selector policy from research.md Decision 3; `data-testid` only when role/label is ambiguous.
- Run `disableAnimations(page)` and `freezeTime(page)` from `tests/support/test-utils.ts` in every visual and time-sensitive test.
- Avoid: real provider creds in fixtures, hardcoded ports outside `tests/support/mock-port.ts`, raw CSS selectors in specs, mocking endpoints that don't exist in `contracts/`.
