# Tasks: UI Model Rate Limit Controls

**Input**: Design documents from `/specs/33-ui-model-rate-limits/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/model-rate-limit-controls.md, quickstart.md

**Tests**: Focused Vitest coverage is included because the specification and quickstart explicitly require validation of action payloads, form controls, inventory display, invalid input, and gateway failure behavior.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing. User Story 1 is the MVP because it lets administrators configure the existing backend `rate_limit_rps` policy from the UI.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm the existing backend/UI contract and test surfaces before implementation.

- [X] T001 Verify the existing backend admin model contract includes `rate_limit_rps` in `pkg/llm/types.go` and `internal/api/admin.go`
- [X] T002 [P] Review current model management UI form and table patterns in `ui/src/components/CreateModelModal.tsx`, `ui/src/components/EditModelModal.tsx`, and `ui/src/app/(dashboard)/models/page.tsx`
- [X] T003 [P] Review existing UI server action test patterns in `ui/src/lib/actions.test.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the shared typed model field required by all stories.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T004 Add `rate_limit_rps: number` to the `Model` interface in `ui/src/lib/types.ts`
- [X] T005 Add shared model rate limit form parsing helper coverage expectations in `ui/src/lib/actions.test.ts`

**Checkpoint**: Foundation ready - user story implementation can now begin.

---

## Phase 3: User Story 1 - Configure Model Rate Limits (Priority: P1) MVP

**Goal**: Administrators can set a non-negative per-model request rate limit when creating or editing a model.

**Independent Test**: Create a model with a positive rate limit, edit that model to a different limit, refresh the model list, and verify the configured value is sent to the gateway and visible in the UI.

### Tests for User Story 1

- [X] T006 [US1] Add `createModel` and `updateModel` payload tests for `rate_limit_rps` in `ui/src/lib/actions.test.ts`
- [X] T007 [P] [US1] Add create modal tests for default model rate limit input and helper copy in `ui/src/components/CreateModelModal.test.tsx`
- [X] T008 [P] [US1] Add edit modal tests for existing model rate limit default value and changed submission in `ui/src/components/EditModelModal.test.tsx`

### Implementation for User Story 1

- [X] T009 [US1] Implement non-negative integer `rate_limit_rps` parsing for model forms in `ui/src/lib/actions.ts`
- [X] T010 [US1] Include `rate_limit_rps` in `createModel` and `updateModel` JSON payloads in `ui/src/lib/actions.ts`
- [X] T011 [US1] Add model rate limit numeric input with default `0` and requests-per-second helper text in `ui/src/components/CreateModelModal.tsx`
- [X] T012 [US1] Add model rate limit numeric input populated from `model.rate_limit_rps` with requests-per-second helper text in `ui/src/components/EditModelModal.tsx`

**Checkpoint**: User Story 1 should be fully functional and testable independently.

---

## Phase 4: User Story 2 - Understand Model Throttle Policy (Priority: P2)

**Goal**: Administrators can scan the model inventory and understand which models have model-level throttling and which have no model-level limit.

**Independent Test**: Load the model inventory with models where `rate_limit_rps` is positive, zero, and missing, then verify each row clearly communicates the model-level policy.

### Tests for User Story 2

- [X] T013 [P] [US2] Add rate limit display tests for positive, zero, and missing model values in `ui/src/components/ModelRateLimitStatus.test.tsx`
- [X] T014 [P] [US2] Add models page table tests for the dedicated model rate limit column and fallback row colspans in `ui/src/app/(dashboard)/models/page.test.tsx`

### Implementation for User Story 2

- [X] T015 [US2] Create `ModelRateLimitStatus` display component for positive and no-limit states in `ui/src/components/ModelRateLimitStatus.tsx`
- [X] T016 [US2] Add a dedicated model rate limit table column and update empty/unavailable row colspans in `ui/src/app/(dashboard)/models/page.tsx`
- [X] T017 [US2] Add inventory copy that distinguishes model-level limits from API key limits in `ui/src/app/(dashboard)/models/page.tsx`

**Checkpoint**: User Stories 1 and 2 should both work independently.

---

## Phase 5: User Story 3 - Handle Invalid Rate Limit Input (Priority: P3)

**Goal**: Administrators receive clear feedback when model rate limit input is invalid or the gateway rejects a create/update request.

**Independent Test**: Attempt to save a negative model rate limit and a gateway-rejected update, then verify the UI keeps the form open, preserves entered state, and displays the failure reason.

### Tests for User Story 3

- [X] T018 [US3] Add negative `rate_limit_rps` rejection tests and gateway error envelope tests in `ui/src/lib/actions.test.ts`
- [X] T019 [P] [US3] Add create modal failure-state tests for invalid rate limits and gateway failures in `ui/src/components/CreateModelModal.test.tsx`
- [X] T020 [P] [US3] Add edit modal failure-state tests for invalid rate limits and gateway failures in `ui/src/components/EditModelModal.test.tsx`

### Implementation for User Story 3

- [X] T021 [US3] Return clear validation errors for negative or non-integer model rate limits in `ui/src/lib/actions.ts`
- [X] T022 [US3] Surface gateway error envelope messages from failed model create and update requests in `ui/src/lib/actions.ts`
- [X] T023 [US3] Ensure create and edit modals keep entered model rate limit state visible after failed saves in `ui/src/components/CreateModelModal.tsx` and `ui/src/components/EditModelModal.tsx`

**Checkpoint**: All user stories should now be independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify the full UI workflow, document any remaining API-only exceptions, and keep existing behavior intact.

- [X] T024 [P] Document that token-per-minute and per-key-per-model limits remain API-only/out of scope in `specs/33-ui-model-rate-limits/quickstart.md`
- [X] T025 Run focused model rate limit UI tests from `ui/package.json` for `ui/src/lib/actions.test.ts`, `ui/src/components/CreateModelModal.test.tsx`, `ui/src/components/EditModelModal.test.tsx`, `ui/src/components/ModelRateLimitStatus.test.tsx`, and `ui/src/app/(dashboard)/models/page.test.tsx`
- [X] T026 Run full UI verification from `ui/package.json` with `npm test -- --run`, `npm run lint`, and `npm run build`
- [ ] T027 Run the manual smoke scenario in `specs/33-ui-model-rate-limits/quickstart.md`
- [X] T028 If backend contract drift was found in T001, run focused Go checks for model admin behavior in `internal/api/admin_test.go` and `internal/storage/`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion and is the MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational completion; can be implemented independently but naturally builds on the same model inventory surface.
- **User Story 3 (Phase 5)**: Depends on Foundational completion; can be implemented independently but should preserve US1/US2 behavior.
- **Polish (Phase 6)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational; no dependency on other stories.
- **User Story 2 (P2)**: Can start after Foundational; no dependency on US1 for display-only behavior if `rate_limit_rps` is typed.
- **User Story 3 (P3)**: Can start after Foundational; validation and failure handling should be reconciled with US1 form/action changes.

### Within Each User Story

- Tests should be written first and fail before implementation.
- Shared TypeScript types should be updated before action/component implementation.
- Server action payload and validation behavior should be implemented before modal integration.
- Table display tests should be in place before adding the inventory column.

### Parallel Opportunities

- T002 and T003 can run in parallel because they inspect different UI areas.
- T007 and T008 can run in parallel because they create separate component test files.
- T013 and T014 can run in parallel because they target component and page tests separately.
- T019 and T020 can run in parallel because they target separate modal test files.
- User Stories 1, 2, and 3 can be staffed in parallel after Phase 2, with coordination around `ui/src/lib/actions.ts`.

---

## Parallel Example: User Story 1

```bash
Task: "Add create modal tests for default model rate limit input and helper copy in ui/src/components/CreateModelModal.test.tsx"
Task: "Add edit modal tests for existing model rate limit default value and changed submission in ui/src/components/EditModelModal.test.tsx"
```

## Parallel Example: User Story 2

```bash
Task: "Add rate limit display tests for positive, zero, and missing model values in ui/src/components/ModelRateLimitStatus.test.tsx"
Task: "Add models page table tests for the dedicated model rate limit column and fallback row colspans in ui/src/app/(dashboard)/models/page.test.tsx"
```

## Parallel Example: User Story 3

```bash
Task: "Add create modal failure-state tests for invalid rate limits and gateway failures in ui/src/components/CreateModelModal.test.tsx"
Task: "Add edit modal failure-state tests for invalid rate limits and gateway failures in ui/src/components/EditModelModal.test.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 for create/edit configuration.
3. Validate that create/update payloads include `rate_limit_rps` and that form controls default to `0`.
4. Stop and demo the create/edit workflow before expanding inventory display and failure handling.

### Incremental Delivery

1. Deliver US1 so administrators can configure limits.
2. Deliver US2 so administrators can inspect model policy at a glance.
3. Deliver US3 so invalid input and gateway failures are clear and state-preserving.
4. Finish Polish tasks and run the quickstart smoke scenario.

### Validation Summary

- Focused UI tests: `cd ui && npm test -- --run src/lib/actions.test.ts`
- Broader UI checks: `cd ui && npm test -- --run && npm run lint && npm run build`
- Backend checks only if contract drift is discovered: `go test ./internal/api/... -run Model -v`
