# Tasks: UI Usage and Audit Dashboard

**Input**: Design documents from `/specs/32-ui-usage-audit-dashboard/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/usage-audit-dashboard.md`, `quickstart.md`

**Tests**: UI tests are required for changed production behavior because this feature closes an admin UI capability gap. Write or update Vitest/React Testing Library coverage for analytics fetch helpers, filter parsing, page rendering, pagination, and empty/unavailable/invalid-filter states. Run Go tests only if backend files are changed.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare the UI surface and confirm the existing backend contracts that the feature consumes.

- [x] T001 Create analytics route and component scaffolding in `ui/src/app/(dashboard)/usage/page.tsx`, `ui/src/app/(dashboard)/usage/loading.tsx`, and `ui/src/components/usage/`
- [x] T002 [P] Verify the existing grouped usage and daily usage response fields in `internal/api/admin_usage.go` against `specs/32-ui-usage-audit-dashboard/contracts/usage-audit-dashboard.md`
- [x] T003 [P] Verify the existing audit query response fields in `internal/api/admin.go` against `specs/32-ui-usage-audit-dashboard/contracts/usage-audit-dashboard.md`
- [x] T004 [P] Review existing UI fetch, formatting, and server-token patterns in `ui/src/lib/actions.ts` and `ui/src/lib/budget.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared analytics contracts, fetch helpers, and URL-backed filter parsing required by every user story.

**Critical**: No user story work can begin until this phase is complete.

- [x] T005 Define TypeScript analytics types for usage summaries, daily usage points, audit records, filters, and fetch states in `ui/src/lib/types.ts`
- [x] T006 Implement authenticated server-side analytics fetch helpers for `/v1/admin/usage`, `/v1/admin/usage/daily`, and `/v1/admin/audit` in `ui/src/lib/analytics.ts`
- [x] T007 Implement URL query parsing and validation for analytics filters, including `from`, `to`, `group_by`, `api_key_id`, `model`, `provider`, `status`, `limit`, and `offset`, in `ui/src/lib/analytics.ts`
- [x] T008 [P] Add Vitest coverage for analytics response parsing, bearer token handling, missing-token unavailable state, 400 invalid-filter mapping, 401/403 auth mapping, and 5xx unavailable mapping in `ui/src/lib/analytics.test.ts`
- [x] T009 [P] Create shared empty, unavailable, and invalid-filter state UI helpers for analytics panels in `ui/src/components/usage/AnalyticsState.tsx`

**Checkpoint**: Foundation ready - user story implementation can now begin in priority order or in parallel by story.

---

## Phase 3: User Story 1 - Review Usage and Spend Summary (Priority: P1) MVP

**Goal**: Administrators can review request counts, token usage, cost, errors, and average latency grouped by model, provider, or API key over a selected time window.

**Independent Test**: Open `/usage`, select each grouping option, adjust the time window, and verify the usage summary updates while preserving submitted filters.

### Tests for User Story 1

- [x] T010 [P] [US1] Add page tests for grouped usage summary rendering, group switching, time-window preservation, empty state, unavailable state, and invalid-filter state in `ui/src/app/(dashboard)/usage/page.test.tsx`
- [x] T011 [P] [US1] Add component tests for usage totals, token counts, error counts, latency display, and small USD formatting in `ui/src/components/usage/UsageSummaryTable.test.tsx`

### Implementation for User Story 1

- [x] T012 [US1] Implement the URL-backed summary and time-window controls in `ui/src/components/usage/UsageFilterForm.tsx`
- [x] T013 [US1] Implement grouped usage summary rendering with cost, tokens, requests, errors, and average latency in `ui/src/components/usage/UsageSummaryTable.tsx`
- [x] T014 [US1] Fetch grouped usage data and render the summary section on the server page in `ui/src/app/(dashboard)/usage/page.tsx`
- [x] T015 [US1] Add the `Usage & Audit` sidebar navigation item linking to `/usage` in `ui/src/components/layout/sidebar-config.tsx`
- [x] T016 [US1] Add a `Usage & Audit` dashboard card linking to `/usage` in `ui/src/app/(dashboard)/page.tsx`

**Checkpoint**: User Story 1 is independently functional and testable as the MVP.

---

## Phase 4: User Story 2 - Inspect Daily Trends (Priority: P2)

**Goal**: Administrators can view daily request, token, and spend trends over a selected time window.

**Independent Test**: Select a multi-day time window and verify the page presents one daily data point per returned day without breaking when days have no usage.

### Tests for User Story 2

- [x] T017 [P] [US2] Add daily trend page tests for multi-day results, empty trends, unavailable state, invalid-filter state, and preservation of shared filters in `ui/src/app/(dashboard)/usage/page.test.tsx`
- [x] T018 [P] [US2] Add component tests for daily date, request count, total token, zero-value, and USD formatting in `ui/src/components/usage/UsageTrendTable.test.tsx`

### Implementation for User Story 2

- [x] T019 [US2] Implement daily usage trend rendering in `ui/src/components/usage/UsageTrendTable.tsx`
- [x] T020 [US2] Render daily trend data from `fetchDailyUsage` alongside the summary section in `ui/src/app/(dashboard)/usage/page.tsx`
- [x] T021 [US2] Preserve shared time-window, API key, and model filters between summary and daily trend queries in `ui/src/lib/analytics.ts`

**Checkpoint**: User Stories 1 and 2 both work independently.

---

## Phase 5: User Story 3 - Search Audit Logs (Priority: P3)

**Goal**: Administrators can search paginated audit records by key, model, provider, status, and time range and inspect investigation fields without exposing secrets.

**Independent Test**: Apply audit filters and pagination in `/usage`, verify matching rows are shown, and inspect request ID, model, provider, status, latency, tokens, cost, streaming, error context, and timestamp.

### Tests for User Story 3

- [x] T022 [P] [US3] Add page tests for audit status filtering, model/provider/key filters, pagination, filter preservation, empty state, unavailable state, and invalid-filter state in `ui/src/app/(dashboard)/usage/page.test.tsx`
- [x] T023 [P] [US3] Add component tests for audit row fields, error row styling, streaming display, pagination controls, and secret-free rendering in `ui/src/components/usage/AuditLogTable.test.tsx`

### Implementation for User Story 3

- [x] T024 [US3] Extend analytics filter controls with audit-specific provider, status, limit, and pagination behavior in `ui/src/components/usage/UsageFilterForm.tsx`
- [x] T025 [US3] Implement paginated audit record rendering with investigation fields and redacted display rules in `ui/src/components/usage/AuditLogTable.tsx`
- [x] T026 [US3] Fetch audit data and render the audit search section on the server page in `ui/src/app/(dashboard)/usage/page.tsx`
- [x] T027 [US3] Reset audit `offset` to `0` when non-pagination filters change in `ui/src/lib/analytics.ts`

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, verification, and final quality checks across all user stories.

- [x] T028 [P] Ensure the `/usage` loading state communicates pending analytics fetches without exposing secrets in `ui/src/app/(dashboard)/usage/loading.tsx`
- [x] T029 [P] Document any remaining analytics or audit capability that is intentionally API-only, or state that none remain, in `specs/32-ui-usage-audit-dashboard/quickstart.md`
- [x] T030 Run focused UI tests with `npm test -- --run` from `ui/`
- [x] T031 Run UI lint and production build checks with `npm run lint` and `npm run build` from `ui/`
- [x] T032 Run `go test ./...` and `golangci-lint run` only if backend Go files were changed; otherwise record that backend code remained read-only in `specs/32-ui-usage-audit-dashboard/quickstart.md`
- [ ] T033 Perform the manual verification flow from `specs/32-ui-usage-audit-dashboard/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - blocks all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational completion - MVP delivery
- **User Story 2 (Phase 4)**: Depends on Foundational completion and can reuse shared filters from US1
- **User Story 3 (Phase 5)**: Depends on Foundational completion and can reuse shared filters from US1
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - no dependency on other stories
- **User Story 2 (P2)**: Can start after Foundational - integrates into the same `/usage` page but remains independently testable
- **User Story 3 (P3)**: Can start after Foundational - integrates into the same `/usage` page but remains independently testable

### Dependency Graph

```text
Phase 1 Setup
  -> Phase 2 Foundational
    -> US1 Review Usage and Spend Summary
    -> US2 Inspect Daily Trends
    -> US3 Search Audit Logs
      -> Phase 6 Polish
```

### Within Each User Story

- Tests should be added before implementation and fail for missing behavior.
- Shared contracts and fetch helpers must be in place before page or component integration.
- Components can be implemented before page integration.
- Each story should pass its independent tests before moving to the next priority.

### Parallel Opportunities

- T002, T003, and T004 can run in parallel during setup.
- T008 and T009 can run in parallel after shared type decisions are clear.
- US1 component tests T010 and T011 can run in parallel.
- US2 tests T017 and T018 can run in parallel.
- US3 tests T022 and T023 can run in parallel.
- After Phase 2, US1, US2, and US3 can be developed by separate contributors with coordination around `ui/src/app/(dashboard)/usage/page.tsx`, `ui/src/components/usage/UsageFilterForm.tsx`, and `ui/src/lib/analytics.ts`.

---

## Parallel Example: User Story 1

```bash
Task: "Add page tests for grouped usage summary rendering, group switching, time-window preservation, empty state, unavailable state, and invalid-filter state in ui/src/app/(dashboard)/usage/page.test.tsx"
Task: "Add component tests for usage totals, token counts, error counts, latency display, and small USD formatting in ui/src/components/usage/UsageSummaryTable.test.tsx"
```

## Parallel Example: User Story 2

```bash
Task: "Add daily trend page tests for multi-day results, empty trends, unavailable state, invalid-filter state, and preservation of shared filters in ui/src/app/(dashboard)/usage/page.test.tsx"
Task: "Add component tests for daily date, request count, total token, zero-value, and USD formatting in ui/src/components/usage/UsageTrendTable.test.tsx"
```

## Parallel Example: User Story 3

```bash
Task: "Add page tests for audit status filtering, model/provider/key filters, pagination, filter preservation, empty state, unavailable state, and invalid-filter state in ui/src/app/(dashboard)/usage/page.test.tsx"
Task: "Add component tests for audit row fields, error row styling, streaming display, pagination controls, and secret-free rendering in ui/src/components/usage/AuditLogTable.test.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. Validate `/usage` grouped summary behavior independently.
5. Stop and demo the MVP before adding trends or audit search.

### Incremental Delivery

1. Add grouped usage and spend summaries.
2. Add daily trend visibility using the same filters.
3. Add audit search and pagination.
4. Run the full UI test, lint, build, and manual verification flow.

### Parallel Team Strategy

1. Complete shared contracts and helpers together.
2. Assign one contributor per story after Phase 2.
3. Coordinate page and filter-form edits because all stories integrate into `/usage`.
4. Merge in priority order: US1, then US2, then US3.

---

## Extension Hooks

**Optional Pre-Hook**: git
Command: `/speckit-git-commit`
Description: Auto-commit before task generation

Prompt: Commit outstanding changes before task generation?
To execute: `/speckit-git-commit`

**Optional Hook**: git
Command: `/speckit-git-commit`
Description: Auto-commit after task generation

Prompt: Commit task changes?
To execute: `/speckit-git-commit`

---

## Notes

- All executable tasks use the required checklist format.
- Backend files are listed for contract verification only; this feature should remain UI-only unless a contract mismatch is discovered.
- Preserve selected filters across empty, unavailable, invalid-filter, and pagination states.
- Never render raw API keys, provider credentials, request payloads, or server-only admin tokens.
