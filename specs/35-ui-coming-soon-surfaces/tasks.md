# Tasks: UI Coming Soon Surfaces

**Input**: Design documents from `specs/35-ui-coming-soon-surfaces/`  
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `quickstart.md`, `contracts/`, `surface-specs/`

**Tests**: Tests are included because the feature specification defines mandatory independent tests, acceptance scenarios, accessibility outcomes, and visual coverage for the eight replacement pages.

**Organization**: Tasks are grouped by user story so each story can be implemented and verified independently after the shared setup and foundational phases.

## Phase 1: Setup

**Purpose**: Establish shared UI contracts and safety utilities needed by every replacement page.

- [X] T001 Review the current placeholder page exports and dashboard shell expectations in `ui/src/app/(dashboard)/logs/page.tsx`, `ui/src/app/(dashboard)/audit/page.tsx`, `ui/src/app/(dashboard)/routes/page.tsx`, `ui/src/app/(dashboard)/guardrails/page.tsx`, `ui/src/app/(dashboard)/teams/page.tsx`, `ui/src/app/(dashboard)/budgets/page.tsx`, `ui/src/app/(dashboard)/rate-limits/page.tsx`, and `ui/src/app/(dashboard)/settings/page.tsx`
- [X] T002 Create shared typed admin surface contracts for log, audit, route, guardrail, team, budget, rate-limit, and settings payloads in `ui/src/lib/admin-surface-contracts.ts`
- [X] T003 [P] Create shared redaction and preview truncation helpers for headers, secret-like keys, prompts, responses, errors, and settings values in `ui/src/lib/redaction.ts`
- [X] T004 [P] Create shared URL query-state parsing and serialization helpers for filter forms and pagination in `ui/src/lib/query-state.ts`
- [X] T005 [P] Add redaction unit coverage for authorization headers, cookies, tokens, provider credentials, raw API keys, prompt previews, response previews, and error summaries in `ui/src/lib/redaction.test.ts`
- [X] T006 [P] Add query-state unit coverage for preserving, clearing, and round-tripping Logs and Audit filters in `ui/src/lib/query-state.test.ts`
- [X] T007 [P] Add shared admin surface formatting helpers for status labels, latency, currency, token counts, percentages, utilization states, and unavailable copy in `ui/src/lib/admin-surface-format.ts`
- [X] T008 [P] Add formatting unit coverage for status labels, budget utilization, tripped labels, and currency display in `ui/src/lib/admin-surface-format.test.ts`

---

## Phase 2: Foundational

**Purpose**: Add deterministic fixtures, mock handlers, and shared UI primitives that block all user stories.

- [X] T009 Create reusable admin surface state components for loading, empty, unavailable, error, and read-only states in `ui/src/components/admin-surface-state.tsx`
- [X] T010 Create reusable admin table, filter bar, drawer, and card layout primitives for the eight surfaces in `ui/src/components/admin-surface-layout.tsx`
- [X] T011 [P] Add route policy fixtures covering fallback, weighted, latency, and single-provider strategies in `ui/tests/fixtures/routes.ts`
- [X] T012 [P] Add settings card fixtures covering Gateway Profile, Security, Notifications, and Display states in `ui/tests/fixtures/settings.ts`
- [X] T013 [P] Add mock route list and mock-only route mutation behavior in `ui/tests/mock/handlers/admin-routes.ts`
- [X] T014 [P] Add mock settings list and local-only display preference behavior in `ui/tests/mock/handlers/admin-settings.ts`
- [X] T015 Register any new route and settings mock handlers with the mock server in `ui/tests/mock/server.ts`
- [X] T016 Extend mock reset state for logs, audit, guardrails, teams, budgets, rate limits, routes, and settings in `ui/tests/mock/state.ts`
- [X] T017 Update mock type exports so E2E handlers and UI contract tests share deterministic page payload shapes in `ui/tests/mock/types.ts`
- [X] T018 Update the navigation or route smoke coverage so all eight replacement routes assert accessible page titles instead of `Coming soon.` in `ui/tests/e2e/navigation.spec.ts`

**Checkpoint**: Shared helpers, fixtures, handlers, and reusable page states are ready. User stories can now proceed independently.

---

## Phase 3: User Story 1 - Investigate Gateway Traffic (Priority: P1)

**Goal**: Gateway operators can use Logs and Audit to investigate request behavior, routing outcomes, failures, spend, and administrative history without raw storage access.

**Independent Test**: Load `/logs` and `/audit` against the deterministic mock backend, filter seeded rows, open a fallback log detail, reload filtered URLs, and verify no secret-bearing headers, credentials, prompts, responses, or errors are exposed.

### Tests for User Story 1

- [X] T019 [P] [US1] Replace Logs Playwright fixme tests with assertions for mixed rows, 5xx filtering, fallback filtering, URL filter restore, empty state, unavailable state, fallback inspector trace, and redacted headers in `ui/tests/e2e/logs.spec.ts`
- [X] T020 [P] [US1] Replace Audit Playwright fixme tests with assertions for actor filtering, action filtering, date filtering, URL filter restore, empty state, unavailable state, newest-first ordering, and redacted error summaries in `ui/tests/e2e/audit.spec.ts`
- [X] T021 [P] [US1] Add component or unit tests for Logs and Audit filter query-state integration in `ui/src/components/logs/logs-filters.test.tsx` and `ui/src/components/audit/audit-filters.test.tsx`

### Implementation for User Story 1

- [X] T022 [P] [US1] Implement Logs data access, typed response normalization, pagination, status/fallback filters, and detail fetch helpers in `ui/src/components/logs/logs-data.ts`
- [X] T023 [P] [US1] Implement Audit data access, typed response normalization, pagination, actor/action/date filters, and newest-first normalization in `ui/src/components/audit/audit-data.ts`
- [X] T024 [US1] Implement the Logs page with accessible title, loading state, table, status/fallback filters, URL query synchronization, pagination, clear-filter empty state, and unavailable state in `ui/src/app/(dashboard)/logs/page.tsx`
- [X] T025 [US1] Implement the request inspector with trace, prompt preview, response preview, redacted headers, fallback stage labeling, and preserved filter state in `ui/src/components/logs/request-inspector.tsx`
- [X] T026 [US1] Implement the Audit page with accessible title, filter form, URL query synchronization, paginated read-only table, status/outcome labels, redacted errors, empty state, and unavailable state in `ui/src/app/(dashboard)/audit/page.tsx`
- [X] T027 [P] [US1] Add Logs surface components for table rows, status badges, provider chain summaries, and fallback-specific accessible labels in `ui/src/components/logs/logs-table.tsx`
- [X] T028 [P] [US1] Add Audit surface components for record rows, cost/token summaries, status labels, and filter controls in `ui/src/components/audit/audit-table.tsx`
- [X] T029 [US1] Verify User Story 1 with focused Playwright runs from `ui/package.json` using `ui/tests/e2e/logs.spec.ts` and `ui/tests/e2e/audit.spec.ts`

**Checkpoint**: Logs and Audit are independently usable for operational investigation and pass their focused functional tests.

---

## Phase 4: User Story 2 - Configure Routing and Safety Controls (Priority: P1)

**Goal**: Gateway administrators can inspect model routing strategies, fallback behavior, and guardrail enabled states while production writes remain unavailable until contracts are reconciled.

**Independent Test**: Load `/routes` and `/guardrails` against seeded providers/models/guardrails, switch route strategy views, toggle `gr_jail` in editable mock-backed context, reload, and confirm persisted state or clear failure behavior.

### Tests for User Story 2

- [X] T030 [P] [US2] Replace Routes Playwright fixme tests with assertions for fallback, weighted, latency, single-provider strategy views, disabled production save controls, route empty state, invalid route validation, and save failure copy in `ui/tests/e2e/routes.spec.ts`
- [X] T031 [P] [US2] Replace Guardrails Playwright fixme tests with assertions for seeded guardrail rows, `gr_jail` toggle success, reload persistence, toggle failure, empty state, unavailable state, and sensitive payload omission in `ui/tests/e2e/guardrails.spec.ts`
- [X] T032 [P] [US2] Add route validation unit tests for missing primary provider, missing fallback provider, negative weights, and zero-total weighted routes in `ui/src/components/routes/route-validation.test.ts`
- [X] T033 [P] [US2] Add guardrail toggle state unit tests for saving, success, failure, and unavailable mutation support in `ui/src/components/guardrails/guardrail-toggle.test.tsx`

### Implementation for User Story 2

- [X] T034 [P] [US2] Implement route data access, typed normalization, and mock-backed mutation gating in `ui/src/components/routes/routes-data.ts`
- [X] T035 [P] [US2] Implement guardrail data access, typed normalization, toggle mutation helpers, and production edit gating in `ui/src/components/guardrails/guardrails-data.ts`
- [X] T036 [US2] Implement the Routes page with accessible title, route list, strategy selection, current policy summary, empty state, unavailable state, and production-unavailable save controls in `ui/src/app/(dashboard)/routes/page.tsx`
- [X] T037 [US2] Implement route diagrams for single, fallback, weighted, and latency strategies with accessible text equivalents and stable strategy selectors in `ui/src/components/routes/route-diagram.tsx`
- [X] T038 [US2] Implement editable route controls with validation for provider mixes, fallback ordering, non-negative weights, and zero-total weighted routes in `ui/src/components/routes/route-editor.tsx`
- [X] T039 [US2] Implement the Guardrails page with accessible title, seeded policy list, enabled state, toggle controls, saving/success/failure feedback, empty state, unavailable state, and sensitive payload omission in `ui/src/app/(dashboard)/guardrails/page.tsx`
- [X] T040 [P] [US2] Add guardrail row and toggle components with `toggle-gr_jail` support and clear unavailable production copy in `ui/src/components/guardrails/guardrails-list.tsx`
- [X] T041 [US2] Verify User Story 2 with focused Playwright runs from `ui/package.json` using `ui/tests/e2e/routes.spec.ts` and `ui/tests/e2e/guardrails.spec.ts`

**Checkpoint**: Routes and Guardrails are independently usable for route/safety inspection, and mock-backed editable states are tested without enabling unsupported production writes.

---

## Phase 5: User Story 3 - Govern Tenants, Budgets, and Rate Limits (Priority: P1)

**Goal**: Gateway administrators can monitor tenant governance, identify near-cap budgets, and find tripped throttling rules across team, key, and model scopes.

**Independent Test**: Load `/teams`, `/budgets`, and `/rate-limits` with seeded fixtures; verify both teams render, Research shows an 85%+ budget warning, Platform does not, and exactly one rate-limit rule shows a tripped state.

### Tests for User Story 3

- [X] T042 [P] [US3] Replace Teams Playwright fixme tests with assertions for Research and Platform rows, Research warning state, Platform non-warning state, empty state, unavailable state, and budget-control guidance in `ui/tests/e2e/teams.spec.ts`
- [X] T043 [P] [US3] Replace Budgets Playwright fixme tests with assertions for Research near-cap state, Platform normal state, edit success in mock-backed context, save failure preservation, key hard-cap context, empty state, and unavailable state in `ui/tests/e2e/budgets.spec.ts`
- [X] T044 [P] [US3] Replace Rate Limits Playwright fixme tests with assertions for seeded rules, exactly one tripped indicator, separate RPS/TPM display, negative value validation, delete success in mock-backed context, unavailable state, and model-level guidance in `ui/tests/e2e/rate-limits.spec.ts`
- [X] T045 [P] [US3] Add budget policy validation unit tests for non-negative limits, non-negative usage, alert threshold bounds, and usage preservation in `ui/src/components/budgets/budget-validation.test.ts`
- [X] T046 [P] [US3] Add rate-limit validation unit tests for scope values, non-negative RPS/TPM, at least one configured limit, and TPM unavailable copy in `ui/src/components/rate-limits/rate-limit-validation.test.ts`

### Implementation for User Story 3

- [X] T047 [P] [US3] Implement Teams data access, typed normalization, budget health mapping, and unavailable service handling in `ui/src/components/teams/teams-data.ts`
- [X] T048 [P] [US3] Implement Budgets data access, typed normalization, utilization calculations, edit helpers, and production edit gating in `ui/src/components/budgets/budgets-data.ts`
- [X] T049 [P] [US3] Implement Rate Limits data access, typed normalization, CRUD helpers, validation helpers, and production edit gating in `ui/src/components/rate-limits/rate-limits-data.ts`
- [X] T050 [US3] Implement the Teams page with accessible title, team grid/table, member counts, budget utilization, near-cap text and visual treatment, empty state, unavailable state, and budget navigation guidance in `ui/src/app/(dashboard)/teams/page.tsx`
- [X] T051 [US3] Implement the Budgets page with accessible title, team/key scope grouping, limit/usage/duration/threshold display, near-cap and over-cap labels, edit controls, save success/failure states, empty state, and unavailable state in `ui/src/app/(dashboard)/budgets/page.tsx`
- [X] T052 [US3] Implement the Rate Limits page with accessible title, scope grouping, RPS/TPM display, tripped text indicator, create/edit/delete controls, validation feedback, TPM availability copy, empty state, and unavailable state in `ui/src/app/(dashboard)/rate-limits/page.tsx`
- [X] T053 [P] [US3] Add team row/card components with `team-research-warn` and accessible utilization labels in `ui/src/components/teams/teams-list.tsx`
- [X] T054 [P] [US3] Add budget policy form and status components that preserve current usage while editing limit, duration, and alert threshold in `ui/src/components/budgets/budget-policy-form.tsx`
- [X] T055 [P] [US3] Add rate-limit rule form and list components with accessible tripped indicators and rule-specific mutation feedback in `ui/src/components/rate-limits/rate-limit-rules.tsx`
- [X] T056 [US3] Verify User Story 3 with focused Playwright runs from `ui/package.json` using `ui/tests/e2e/teams.spec.ts`, `ui/tests/e2e/budgets.spec.ts`, and `ui/tests/e2e/rate-limits.spec.ts`

**Checkpoint**: Teams, Budgets, and Rate Limits are independently usable for spend and throttling governance, including warning and tripped states.

---

## Phase 6: User Story 4 - Manage Organization-Level Settings (Priority: P2)

**Goal**: Gateway administrators can review bounded organization settings across Gateway Profile, Security, Notifications, and Display, with unavailable backend-dependent controls clearly marked.

**Independent Test**: Load `/settings` and verify all four settings cards render with clear save states, validation, redaction, and disabled/unavailable behavior for backend-dependent controls.

### Tests for User Story 4

- [X] T057 [P] [US4] Replace Settings Playwright fixme tests with assertions for the four required cards, read-only/unavailable backend controls, local-only display preference save success, validation failure, save failure, and redacted secret-like values in `ui/tests/e2e/settings.spec.ts`
- [X] T058 [P] [US4] Add settings card validation unit tests for dirty, saving, success, validation error, failure, unavailable, and secret redaction states in `ui/src/components/settings/settings-card.test.tsx`

### Implementation for User Story 4

- [X] T059 [P] [US4] Implement settings data access, typed card normalization, local-only display preference persistence, and production edit gating in `ui/src/components/settings/settings-data.ts`
- [X] T060 [US4] Implement the Settings page with accessible title, Gateway Profile, Security, Notifications, Display cards, read-only/unavailable copy, safe display preference editing, validation, save feedback, and redaction in `ui/src/app/(dashboard)/settings/page.tsx`
- [X] T061 [US4] Implement reusable settings card and field components with dirty, saving, success, validation error, failure, unavailable, read-only, and redacted states in `ui/src/components/settings/settings-card.tsx`
- [X] T062 [US4] Verify User Story 4 with focused Playwright runs from `ui/package.json` using `ui/tests/e2e/settings.spec.ts`

**Checkpoint**: Settings is independently usable as a bounded organization-level settings surface without exposing unsupported production writes.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Validate accessibility, visual coverage, documentation, and full-feature confidence after all independently testable stories are complete.

- [X] T063 [P] Add or update accessibility coverage for all eight replacement routes and remove stale stub-route assumptions in `ui/tests/e2e/a11y.spec.ts`
- [X] T064 [P] Add visual snapshots for Logs mixed/fallback/empty, request inspector states, Audit populated, Routes strategy variants, Guardrails mixed state, Teams populated, Budgets near/over cap, Rate Limits tripped, and Settings all four cards in `ui/tests/e2e/visual.spec.ts`
- [X] T065 [P] Update the admin UI testing documentation to describe the eight replacement surfaces, selectors, a11y expectations, visual snapshots, and focused run commands in `ui/tests/README.md`
- [X] T066 [P] Update high-level testing expectations for the shipped surface states and any remaining API-only capabilities in `TESTING.md`
- [X] T067 Verify UI lint and unit tests from `ui/package.json` using `npm run lint` and `npm run test`
- [X] T068 Verify focused E2E coverage from `ui/package.json` using `npm run test:e2e -- tests/e2e/logs.spec.ts tests/e2e/audit.spec.ts tests/e2e/routes.spec.ts tests/e2e/guardrails.spec.ts tests/e2e/teams.spec.ts tests/e2e/budgets.spec.ts tests/e2e/rate-limits.spec.ts tests/e2e/settings.spec.ts`
- [X] T069 Verify full UI E2E confidence from `ui/package.json` using `npm run test:e2e`
- [X] T070 Document any remaining production API gaps and follow-up triggers for Routes, Guardrails, Budgets, Rate Limits, Teams, and Settings in `specs/35-ui-coming-soon-surfaces/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup**: No dependencies.
- **Phase 2: Foundational**: Depends on Phase 1 shared contracts and utilities.
- **Phase 3: User Story 1**: Depends on Phase 2; can ship as MVP.
- **Phase 4: User Story 2**: Depends on Phase 2; independent from User Stories 1, 3, and 4 after shared foundations.
- **Phase 5: User Story 3**: Depends on Phase 2; independent from User Stories 1, 2, and 4 after shared foundations.
- **Phase 6: User Story 4**: Depends on Phase 2; independent from User Stories 1, 2, and 3 after shared foundations.
- **Phase 7: Polish**: Depends on completed story phases.

### Story Completion Order

1. **MVP**: User Story 1, because Logs and Audit provide the highest-value operational investigation surface.
2. **Parallel P1 expansion**: User Story 2 and User Story 3 can proceed in parallel once shared foundations are complete.
3. **P2 completion**: User Story 4 follows shared foundations and can run in parallel if capacity is available.

---

## Parallel Execution Examples

### User Story 1

```text
Task: T019 in ui/tests/e2e/logs.spec.ts
Task: T020 in ui/tests/e2e/audit.spec.ts
Task: T022 in ui/src/components/logs/logs-data.ts
Task: T023 in ui/src/components/audit/audit-data.ts
```

### User Story 2

```text
Task: T030 in ui/tests/e2e/routes.spec.ts
Task: T031 in ui/tests/e2e/guardrails.spec.ts
Task: T034 in ui/src/components/routes/routes-data.ts
Task: T035 in ui/src/components/guardrails/guardrails-data.ts
```

### User Story 3

```text
Task: T042 in ui/tests/e2e/teams.spec.ts
Task: T043 in ui/tests/e2e/budgets.spec.ts
Task: T044 in ui/tests/e2e/rate-limits.spec.ts
Task: T047 in ui/src/components/teams/teams-data.ts
Task: T048 in ui/src/components/budgets/budgets-data.ts
Task: T049 in ui/src/components/rate-limits/rate-limits-data.ts
```

### User Story 4

```text
Task: T057 in ui/tests/e2e/settings.spec.ts
Task: T058 in ui/src/components/settings/settings-card.test.tsx
Task: T059 in ui/src/components/settings/settings-data.ts
```

---

## Implementation Strategy

### MVP First

Complete Phase 1 and Phase 2, then deliver Phase 3 for Logs and Audit. This satisfies the highest-priority operational investigation workflow and proves the shared redaction, query-state, loading, empty, unavailable, and E2E patterns before expanding to other pages.

### Incremental Delivery

1. Deliver User Story 1 and verify focused Logs/Audit E2E.
2. Deliver User Story 2 for routing and safety controls with production writes gated.
3. Deliver User Story 3 for tenant, budget, and throttling governance.
4. Deliver User Story 4 for bounded organization settings.
5. Run cross-cutting a11y, visual, lint, unit, and full E2E checks.

### Team Parallelization

After Phase 2, separate agents or engineers can work by surface group because Logs/Audit, Routes/Guardrails, Teams/Budgets/Rate Limits, and Settings touch distinct page/component/test files and only share completed helpers.

---

## Validation Summary

- **Total tasks**: 70
- **Setup tasks**: 8
- **Foundational tasks**: 10
- **User Story 1 tasks**: 11
- **User Story 2 tasks**: 12
- **User Story 3 tasks**: 15
- **User Story 4 tasks**: 6
- **Polish tasks**: 8
- **Parallel opportunities**: 36 tasks marked `[P]`
- **Suggested MVP scope**: Phase 1, Phase 2, and Phase 3 User Story 1
- **Format validation**: All task checklist entries use `- [X] T###`, include `[P]` only for parallelizable work, include `[US#]` only for user-story phases, and include explicit file paths.
