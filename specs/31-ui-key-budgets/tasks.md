# Tasks: UI API Key Budget Controls

**Input**: Design documents from `/workspaces/LLMGopher/specs/31-ui-key-budgets/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/ui_key_budget_contract.md`, `quickstart.md`

**Tests**: UI behavior changes require focused Vitest coverage plus lint. Contract behavior is exercised through server action tests that mock gateway responses; Go tests are only required if backend budget handlers or auth behavior change.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested independently after the shared foundation is complete.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare local UI configuration and documentation for server-side budget endpoint auth.

- [X] T001 Document `LLMGOPHER_UI_ADMIN_API_KEY` setup and budget validation commands in `/workspaces/LLMGopher/ui/README.md`
- [X] T002 Add server-only `LLMGOPHER_UI_ADMIN_API_KEY` environment wiring for the `ui` service in `/workspaces/LLMGopher/docker-compose.yaml`
- [X] T003 [P] Add budget feature quickstart notes and no-raw-API-call validation expectations in `/workspaces/LLMGopher/specs/31-ui-key-budgets/quickstart.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared budget contracts, validation helpers, and protected gateway action plumbing used by every story.

**CRITICAL**: No user story work should begin until this phase is complete.

- [X] T004 [P] Add `APIKeyBudget`, `APIKeyBudgetState`, `APIKeyBudgetFormValues`, and `BudgetStatusIndicator` types in `/workspaces/LLMGopher/ui/src/lib/types.ts`
- [X] T005 [P] Implement budget response parsing, status derivation, currency/percentage formatting, and form validation helpers in `/workspaces/LLMGopher/ui/src/lib/budget.ts`
- [X] T006 [P] Add unit tests for budget status derivation, invalid gateway values, and form validation rules in `/workspaces/LLMGopher/ui/src/lib/budget.test.ts`
- [X] T007 Add server-only admin bearer token helper and budget endpoint URL helpers in `/workspaces/LLMGopher/ui/src/lib/actions.ts`
- [X] T008 Add action tests for protected budget requests, missing admin token handling, 401/403 errors, and OpenAI-compatible error envelope parsing in `/workspaces/LLMGopher/ui/src/lib/actions.test.ts`

**Checkpoint**: Budget contracts and protected server action infrastructure are ready.

---

## Phase 3: User Story 1 - View Key Budget State (Priority: P1) MVP

**Goal**: Administrators can inspect a key's budget limit, spend, remaining balance, threshold, duration, reset timing, and no-budget/unavailable states from `/keys`.

**Independent Test**: Mock key and budget gateway responses, render `/keys`, and verify configured, unbudgeted, and unavailable budget states display without exposing admin credentials.

### Tests for User Story 1

- [X] T009 [P] [US1] Add server action tests for `fetchAPIKeyBudget` success, 404 unbudgeted, and unavailable/auth failure states in `/workspaces/LLMGopher/ui/src/lib/actions.test.ts`
- [X] T010 [P] [US1] Add budget status component tests for configured, unbudgeted, near-threshold, exhausted, and unavailable states in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetStatus.test.tsx`
- [X] T011 [P] [US1] Extend `/keys` page tests for budget column rendering, expanded empty/unavailable column spans, and per-key budget fetch behavior in `/workspaces/LLMGopher/ui/src/app/(dashboard)/keys/page.test.tsx`

### Implementation for User Story 1

- [X] T012 [US1] Implement `fetchAPIKeyBudget(apiKeyID)` using server-only bearer auth and 404-to-unbudgeted mapping in `/workspaces/LLMGopher/ui/src/lib/actions.ts`
- [X] T013 [P] [US1] Create `APIKeyBudgetStatus` to render spent, remaining, threshold, duration, reset time, and setup/unavailable copy in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetStatus.tsx`
- [X] T014 [US1] Fetch budget state for visible keys without unbounded client-side fan-out in `/workspaces/LLMGopher/ui/src/app/(dashboard)/keys/page.tsx`
- [X] T015 [US1] Add a Budget table column and pass each key's budget state into row UI in `/workspaces/LLMGopher/ui/src/app/(dashboard)/keys/page.tsx`
- [X] T016 [US1] Add a read/manage budget entry point beside existing key actions without rendering raw tokens in `/workspaces/LLMGopher/ui/src/components/APIKeyRowActions.tsx`
- [X] T017 [US1] Update row action tests to verify the budget entry point exists and does not expose admin credentials in `/workspaces/LLMGopher/ui/src/components/APIKeyRowActions.test.tsx`

**Checkpoint**: User Story 1 is fully functional and testable independently.

---

## Phase 4: User Story 2 - Set or Update Budgets (Priority: P2)

**Goal**: Administrators can create or update key budgets, alert thresholds, duration, and reset timing while preserving current spend unless reset is explicitly invoked.

**Independent Test**: Submit valid and invalid budget forms through the UI, verify validation feedback, gateway `PUT` payloads, successful refresh, and preserved spend display.

### Tests for User Story 2

- [X] T018 [P] [US2] Add server action tests for `upsertAPIKeyBudget` payload validation, gateway `PUT`, revalidation, and error messages in `/workspaces/LLMGopher/ui/src/lib/actions.test.ts`
- [X] T019 [P] [US2] Add budget form tests for valid create/update submissions, invalid limit, invalid threshold, invalid duration, and missing reset time in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetForm.test.tsx`
- [X] T020 [P] [US2] Add budget modal tests for opening from a row, pre-filling existing budget values, successful save toasts, and refresh behavior in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.test.tsx`

### Implementation for User Story 2

- [X] T021 [US2] Implement `upsertAPIKeyBudget(apiKeyID, formData)` with helper validation, gateway `PUT`, error parsing, and `/keys` revalidation in `/workspaces/LLMGopher/ui/src/lib/actions.ts`
- [X] T022 [P] [US2] Create `APIKeyBudgetForm` with budget limit, alert threshold, duration, reset time fields, and field-specific validation messages in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetForm.tsx`
- [X] T023 [US2] Create `APIKeyBudgetModal` to show current state, host `APIKeyBudgetForm`, save changes, show success/error toasts, and refresh the router after mutations in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.tsx`
- [X] T024 [US2] Wire `APIKeyBudgetModal` into `APIKeyRowActions` with current budget state and existing edit/delete actions preserved in `/workspaces/LLMGopher/ui/src/components/APIKeyRowActions.tsx`
- [X] T025 [US2] Update `/keys` page integration to pass budget state into row actions and keep existing key lifecycle flows unchanged in `/workspaces/LLMGopher/ui/src/app/(dashboard)/keys/page.tsx`

**Checkpoint**: User Stories 1 and 2 both work independently.

---

## Phase 5: User Story 3 - Reset or Remove Budgets (Priority: P3)

**Goal**: Administrators can reset budget spend counters or remove budget limits with explicit confirmation and visible refreshed state.

**Independent Test**: Invoke reset and remove flows for a budgeted key, verify confirmation dialogs are required, gateway calls are made, and the UI reflects zero spend or the unbudgeted state after refresh.

### Tests for User Story 3

- [X] T026 [P] [US3] Add server action tests for `resetAPIKeyBudget` and `deleteAPIKeyBudget` success, revalidation, and gateway error handling in `/workspaces/LLMGopher/ui/src/lib/actions.test.ts`
- [X] T027 [P] [US3] Add modal tests proving reset requires confirmation, shows zero-spend state after success, and handles delayed refresh messaging in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.test.tsx`
- [X] T028 [P] [US3] Add modal tests proving remove requires confirmation, calls delete, and returns the UI to the no-budget state in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.test.tsx`

### Implementation for User Story 3

- [X] T029 [US3] Implement `resetAPIKeyBudget(apiKeyID)` with gateway `POST /budget/reset`, error parsing, and `/keys` revalidation in `/workspaces/LLMGopher/ui/src/lib/actions.ts`
- [X] T030 [US3] Implement `deleteAPIKeyBudget(apiKeyID)` with gateway `DELETE /budget`, 204 handling, error parsing, and `/keys` revalidation in `/workspaces/LLMGopher/ui/src/lib/actions.ts`
- [X] T031 [US3] Add reset confirmation dialog and loading/refresh feedback to `APIKeyBudgetModal` in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.tsx`
- [X] T032 [US3] Add remove confirmation dialog and unbudgeted-state transition handling to `APIKeyBudgetModal` in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.tsx`
- [X] T033 [US3] Ensure reset and remove controls are disabled for unbudgeted or unavailable states in `/workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.tsx`

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, documentation, and regression protection across the full budget lifecycle.

- [X] T034 [P] Update UI documentation for budget lifecycle behavior, server-only admin token handling, and troubleshooting in `/workspaces/LLMGopher/ui/README.md`
- [X] T035 [P] Update feature quickstart expected results for configured, unbudgeted, reset, remove, invalid, and auth-unavailable flows in `/workspaces/LLMGopher/specs/31-ui-key-budgets/quickstart.md`
- [X] T036 Add regression coverage proving existing create/edit/delete/deactivate key lifecycle controls still work with budget controls present in `/workspaces/LLMGopher/ui/src/components/APIKeyRowActions.test.tsx`
- [X] T037 Run `npm test` and fix any failures from budget action, helper, page, modal, or row action tests in `/workspaces/LLMGopher/ui/package.json`
- [X] T038 Run `npm run lint` and fix lint issues in `/workspaces/LLMGopher/ui/package.json`
- [ ] T039 Run the manual quickstart validation against `make dev` using `LLMGOPHER_UI_ADMIN_API_KEY=sk-test-key-1` and record any deviations in `/workspaces/LLMGopher/specs/31-ui-key-budgets/quickstart.md`
- [ ] T040 If backend auth or handler behavior changes during implementation, run focused budget API tests and document the command result in `/workspaces/LLMGopher/specs/31-ui-key-budgets/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion; delivers the MVP view/read surface.
- **User Story 2 (Phase 4)**: Depends on Foundational completion and benefits from US1 display state, but `upsertAPIKeyBudget` and form tests can be developed independently after Phase 2.
- **User Story 3 (Phase 5)**: Depends on Foundational completion and budget modal/action patterns from US2.
- **Polish (Phase 6)**: Depends on the desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on other user stories after Phase 2.
- **US2 (P2)**: Can begin after Phase 2; final UI wiring expects the budget status/modal entry point from US1.
- **US3 (P3)**: Can begin server action tests after Phase 2; final modal controls expect US2 modal structure.

### Within Each User Story

- Write listed tests before implementation and confirm they fail for the missing behavior.
- Implement shared helpers before server actions.
- Implement server actions before client components that call them.
- Wire `/keys` page integration after components and actions are available.
- Validate each story independently before moving to the next priority.

### Parallel Opportunities

- Setup documentation task T003 can run alongside T001 and T002.
- Foundational type/helper/test tasks T004, T005, and T006 can run in parallel before T007 and T008.
- US1 tests T009, T010, and T011 can run in parallel; component T013 can run while action T012 is implemented.
- US2 tests T018, T019, and T020 can run in parallel; form T022 can run while action T021 is implemented.
- US3 tests T026, T027, and T028 can run in parallel before reset/remove implementation.
- Polish documentation T034 and T035 can run in parallel with final regression test updates T036.

---

## Parallel Example: User Story 1

```bash
Task: "Add server action tests for fetchAPIKeyBudget success, 404 unbudgeted, and unavailable/auth failure states in /workspaces/LLMGopher/ui/src/lib/actions.test.ts"
Task: "Add budget status component tests for configured, unbudgeted, near-threshold, exhausted, and unavailable states in /workspaces/LLMGopher/ui/src/components/APIKeyBudgetStatus.test.tsx"
Task: "Extend /keys page tests for budget column rendering, expanded empty/unavailable column spans, and per-key budget fetch behavior in /workspaces/LLMGopher/ui/src/app/(dashboard)/keys/page.test.tsx"
```

## Parallel Example: User Story 2

```bash
Task: "Add server action tests for upsertAPIKeyBudget payload validation, gateway PUT, revalidation, and error messages in /workspaces/LLMGopher/ui/src/lib/actions.test.ts"
Task: "Add budget form tests for valid create/update submissions, invalid limit, invalid threshold, invalid duration, and missing reset time in /workspaces/LLMGopher/ui/src/components/APIKeyBudgetForm.test.tsx"
Task: "Create APIKeyBudgetForm with budget limit, alert threshold, duration, reset time fields, and field-specific validation messages in /workspaces/LLMGopher/ui/src/components/APIKeyBudgetForm.tsx"
```

## Parallel Example: User Story 3

```bash
Task: "Add server action tests for resetAPIKeyBudget and deleteAPIKeyBudget success, revalidation, and gateway error handling in /workspaces/LLMGopher/ui/src/lib/actions.test.ts"
Task: "Add modal tests proving reset requires confirmation, shows zero-spend state after success, and handles delayed refresh messaging in /workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.test.tsx"
Task: "Add modal tests proving remove requires confirmation, calls delete, and returns the UI to the no-budget state in /workspaces/LLMGopher/ui/src/components/APIKeyBudgetModal.test.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup and Phase 2 shared contracts/actions.
2. Complete Phase 3 to display configured, unbudgeted, and unavailable budget states on `/keys`.
3. Stop and validate with focused tests plus manual view-only quickstart steps.

### Incremental Delivery

1. Deliver US1 budget visibility without enabling mutation controls.
2. Add US2 create/update controls and validate spend preservation.
3. Add US3 reset/remove controls with explicit confirmation.
4. Run full UI tests, lint, and manual quickstart validation.

### Parallel Team Strategy

1. One developer owns shared TypeScript budget helpers and action tests.
2. One developer owns `/keys` page and status rendering.
3. One developer owns modal/form mutation flows once shared action contracts are available.

---

## Notes

- `[P]` tasks touch different files or independent test scopes and can run in parallel.
- `[US1]`, `[US2]`, and `[US3]` map directly to the prioritized user stories in `spec.md`.
- The gateway budget API remains the source of truth; do not add a UI-only aggregation endpoint unless implementation evidence shows it is required.
- Keep `LLMGOPHER_UI_ADMIN_API_KEY` server-only and never pass it to client components, browser state, logs, or rendered markup.
- Treat gateway `404` from budget lookup as an unbudgeted state, not a page-level failure.
