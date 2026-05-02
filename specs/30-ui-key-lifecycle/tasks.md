# Tasks: UI API Key Lifecycle Controls

**Input**: Design documents from `/specs/30-ui-key-lifecycle/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/ui_key_lifecycle_contract.md`, `quickstart.md`

**Tests**: UI behavior changes require focused tests. This feature adds a minimal UI test harness because `ui/` currently has lint/build scripts but no test script.

**Organization**: Tasks are grouped by user story so each story can be implemented and verified independently after shared foundations are complete.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel with other marked tasks in the same phase because it touches different files or does not depend on incomplete tasks.
- **[Story]**: User-story tasks are labeled `[US1]`, `[US2]`, or `[US3]`.
- Every task includes an exact file path.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the UI test harness and supporting component inventory before lifecycle UI work begins.

- [X] T001 Add Vitest, React Testing Library, jsdom, and `npm test` script in `ui/package.json` and `ui/package-lock.json`
- [X] T002 [P] Create Vitest configuration for Next.js path aliases and jsdom in `ui/vitest.config.ts`
- [X] T003 [P] Create React Testing Library setup for cleanup and jest-dom matchers in `ui/src/test/setup.ts`
- [X] T004 [P] Add reusable textarea UI primitive for metadata JSON input in `ui/src/components/ui/textarea.tsx`
- [X] T005 [P] Add reusable checkbox UI primitive for model allowlist controls in `ui/src/components/ui/checkbox.tsx`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared types, action helpers, and model data plumbing required before any user story can be implemented.

**Critical**: No user story work can begin until this phase is complete.

- [X] T006 Expand `APIKey` and `Model` TypeScript contracts with lifecycle fields in `ui/src/lib/types.ts`
- [X] T007 [P] Add `APIKeyFormValues`, `APIKeyMutationResult`, and gateway error envelope types in `ui/src/lib/types.ts`
- [X] T008 Implement shared API key form parsing and metadata JSON validation helpers in `ui/src/lib/actions.ts`
- [X] T009 Implement gateway error extraction for nested OpenAI-compatible error envelopes in `ui/src/lib/actions.ts`
- [X] T010 Implement `fetchModelsForKeyForms` server helper using `cache: "no-store"` in `ui/src/lib/actions.ts`
- [X] T011 Update the keys page to fetch models alongside keys and pass model choices to key UI components in `ui/src/app/(dashboard)/keys/page.tsx`
- [X] T012 [P] Create reusable model allowlist label/selection utilities that preserve stale identifiers in `ui/src/lib/key-lifecycle.ts`
- [X] T013 [P] Add unit tests for API key form parsing, metadata validation, and gateway error extraction in `ui/src/lib/actions.test.ts`
- [X] T014 [P] Add unit tests for model allowlist label/selection utilities in `ui/src/lib/key-lifecycle.test.ts`

**Checkpoint**: Shared lifecycle contracts and helpers are ready; user story implementation can proceed.

---

## Phase 3: User Story 1 - Update Existing Keys (Priority: P1) - MVP

**Goal**: Administrators can view and edit key name, rate limit, expiration, metadata, allowed models, and active status from `/keys`.

**Independent Test**: Create or select a key in the UI, edit every mutable field, save, refresh the inventory, and verify the saved values remain visible without exposing raw key material.

### Tests for User Story 1

- [X] T015 [P] [US1] Add server-action tests for successful API key create/update payloads including expiration, metadata, allowed models, and `is_active` in `ui/src/lib/actions.test.ts`
- [X] T016 [P] [US1] Add server-action tests for invalid metadata, negative rate limit, gateway-unavailable, and not-found update errors in `ui/src/lib/actions.test.ts`
- [X] T017 [P] [US1] Add component tests for create/edit lifecycle forms preserving form state and one-time raw key display in `ui/src/components/APIKeyLifecycleForm.test.tsx`
- [X] T018 [P] [US1] Add component tests for key inventory lifecycle field rendering and stale model labels in `ui/src/app/(dashboard)/keys/page.test.tsx`

### Implementation for User Story 1

- [X] T019 [US1] Extend `createAPIKey` to submit expiration, metadata, and allowed models while preserving one-time raw key return in `ui/src/lib/actions.ts`
- [X] T020 [US1] Add `updateAPIKey` server action with validation, gateway error context, and `revalidatePath("/keys")` in `ui/src/lib/actions.ts`
- [X] T021 [P] [US1] Create reusable API key lifecycle form fields for name, rate limit, expiration, metadata JSON, and model allowlist in `ui/src/components/APIKeyLifecycleForm.tsx`
- [X] T022 [US1] Update create modal to use shared lifecycle fields, model choices, and raw-key clearing on close in `ui/src/components/CreateAPIKeyModal.tsx`
- [X] T023 [P] [US1] Create edit modal for existing API keys with initial values and preserved form state on failure in `ui/src/components/EditAPIKeyModal.tsx`
- [X] T024 [US1] Create key row actions entry point that opens the edit modal in `ui/src/components/APIKeyRowActions.tsx`
- [X] T025 [US1] Expand the key inventory table with active, expired, unrestricted/restricted, metadata, expiration, and actions columns in `ui/src/app/(dashboard)/keys/page.tsx`
- [X] T026 [US1] Add action-specific success/error toasts and delayed refresh behavior after create/update in `ui/src/components/CreateAPIKeyModal.tsx`
- [X] T027 [US1] Add update success/error toasts and delayed refresh behavior in `ui/src/components/EditAPIKeyModal.tsx`
- [X] T028 [US1] Document that no key lifecycle capability remains API-only and rotation is out of scope in `specs/30-ui-key-lifecycle/quickstart.md`

**Checkpoint**: User Story 1 is complete when all lifecycle fields are visible, editable, persisted after refresh, and covered by focused UI tests.

---

## Phase 4: User Story 2 - Disable or Re-enable Keys (Priority: P2)

**Goal**: Administrators can deactivate and reactivate keys from the UI without creating replacements.

**Independent Test**: Toggle one key inactive and active again from `/keys`, then verify the same key ID remains visible and the status/action labels change correctly.

### Tests for User Story 2

- [X] T029 [P] [US2] Add server-action tests for deactivate/reactivate updates using `is_active` in `ui/src/lib/actions.test.ts`
- [X] T030 [P] [US2] Add component tests for active and inactive row action labels, disabled pending state, and status badge changes in `ui/src/components/APIKeyRowActions.test.tsx`

### Implementation for User Story 2

- [X] T031 [US2] Add `setAPIKeyActiveState` server action wrapping `PUT /v1/admin/keys/{id}` with `is_active` payloads in `ui/src/lib/actions.ts`
- [X] T032 [US2] Add deactivate/reactivate dropdown actions, pending state, and action-specific toasts in `ui/src/components/APIKeyRowActions.tsx`
- [X] T033 [US2] Ensure the keys table renders reversible active/inactive controls and delayed refresh feedback in `ui/src/app/(dashboard)/keys/page.tsx`
- [X] T034 [US2] Add deactivate/reactivate smoke-check notes and expected cache-sync behavior in `specs/30-ui-key-lifecycle/quickstart.md`

**Checkpoint**: User Story 2 is complete when status toggles work on the same key ID and failures preserve the current inventory state.

---

## Phase 5: User Story 3 - Delete Keys (Priority: P3)

**Goal**: Administrators can permanently delete keys only after explicit confirmation and see synchronization feedback.

**Independent Test**: Delete a non-production test key from `/keys`, confirm the warning, and verify the key disappears after refresh or the UI explains a longer synchronization wait.

### Tests for User Story 3

- [X] T035 [P] [US3] Add server-action tests for successful delete, failed delete, and deletion sync timeout in `ui/src/lib/actions.test.ts`
- [X] T036 [P] [US3] Add component tests for destructive confirmation copy, cancel path, and delete pending state in `ui/src/components/APIKeyRowActions.test.tsx`

### Implementation for User Story 3

- [X] T037 [US3] Add `deleteAPIKey` server action and `waitForAPIKeyDeletionSync` polling helper in `ui/src/lib/actions.ts`
- [X] T038 [US3] Add destructive delete confirmation dialog naming the key and explaining permanence in `ui/src/components/APIKeyRowActions.tsx`
- [X] T039 [US3] Add post-delete refresh, sync waiting state, timeout warning, and failure toasts in `ui/src/components/APIKeyRowActions.tsx`
- [X] T040 [US3] Ensure deleted-key empty and unavailable states have correct table column spans after the actions column is added in `ui/src/app/(dashboard)/keys/page.tsx`
- [X] T041 [US3] Add delete smoke-check notes and irreversible-action warning criteria in `specs/30-ui-key-lifecycle/quickstart.md`

**Checkpoint**: User Story 3 is complete when delete requires confirmation 100% of the time and removes the key from the refreshed inventory.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation, and consistency work across all lifecycle flows.

- [X] T042 [P] Update lifecycle contract notes with any final UI behavior details in `specs/30-ui-key-lifecycle/contracts/ui_key_lifecycle_contract.md`
- [X] T043 [P] Update feature plan verification notes if the implemented test harness differs from the original plan in `specs/30-ui-key-lifecycle/plan.md`
- [X] T044 [P] Verify no raw API key material is rendered outside create success state or included in toast/error paths in `ui/src/components/CreateAPIKeyModal.tsx`
- [X] T045 [P] Verify no raw API key material is rendered in edit, row actions, table cells, or tests in `ui/src/components/EditAPIKeyModal.tsx`
- [X] T046 Run `npm test` from `ui/` and fix any failures in `ui/src/lib/actions.test.ts`
- [X] T047 Run `npm run lint` from `ui/` and fix any lint failures in `ui/src/app/(dashboard)/keys/page.tsx`
- [X] T048 Run `npm run build` from `ui/` and fix any build failures in `ui/src/lib/types.ts`
- [ ] T049 Run quickstart smoke checks against the Docker Compose stack and record any deviations in `specs/30-ui-key-lifecycle/quickstart.md`
- [X] T050 Run `go test ./internal/api/... -run 'Test.*APIKey' -v` only if backend key contracts changed, and record the result in `specs/30-ui-key-lifecycle/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; can start immediately.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational; MVP scope.
- **User Story 2 (Phase 4)**: Depends on Foundational and benefits from the row actions created in US1.
- **User Story 3 (Phase 5)**: Depends on Foundational and the row actions entry point created in US1.
- **Polish (Phase 6)**: Depends on all selected user stories.

### User Story Dependencies

- **US1 (P1)**: First deliverable; no dependency on US2 or US3.
- **US2 (P2)**: Uses the row action structure from US1 but remains independently testable through `is_active` updates.
- **US3 (P3)**: Uses the row action structure from US1 but remains independently testable through delete confirmation and sync polling.

### Within Each User Story

- Tests are written first and should fail before implementation.
- Server actions and validation helpers should land before client components that call them.
- Component changes should land before page integration.
- Each checkpoint should be validated before moving to the next story.

### Parallel Opportunities

- T002-T005 can run in parallel after T001 is understood.
- T012-T014 can run in parallel with T008-T011 once shared type shapes are clear.
- US1 tests T015-T018 can run in parallel before implementation.
- US1 component tasks T021 and T023 can run in parallel after T019-T020.
- US2 tests T029-T030 can run in parallel with US2 implementation planning.
- US3 tests T035-T036 can run in parallel with US3 implementation planning.
- Polish documentation tasks T042-T043 can run in parallel with security review tasks T044-T045.

---

## Parallel Example: User Story 1

```bash
Task: "Add server-action tests for successful API key create/update payloads including expiration, metadata, allowed models, and is_active in ui/src/lib/actions.test.ts"
Task: "Add component tests for create/edit lifecycle forms preserving form state and one-time raw key display in ui/src/components/APIKeyLifecycleForm.test.tsx"
Task: "Add component tests for key inventory lifecycle field rendering and stale model labels in ui/src/app/(dashboard)/keys/page.test.tsx"
```

```bash
Task: "Create reusable API key lifecycle form fields for name, rate limit, expiration, metadata JSON, and model allowlist in ui/src/components/APIKeyLifecycleForm.tsx"
Task: "Create edit modal for existing API keys with initial values and preserved form state on failure in ui/src/components/EditAPIKeyModal.tsx"
```

## Parallel Example: User Story 2

```bash
Task: "Add server-action tests for deactivate/reactivate updates using is_active in ui/src/lib/actions.test.ts"
Task: "Add component tests for active and inactive row action labels, disabled pending state, and status badge changes in ui/src/components/APIKeyRowActions.test.tsx"
```

## Parallel Example: User Story 3

```bash
Task: "Add server-action tests for successful delete, failed delete, and deletion sync timeout in ui/src/lib/actions.test.ts"
Task: "Add component tests for destructive confirmation copy, cancel path, and delete pending state in ui/src/components/APIKeyRowActions.test.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 to establish repeatable UI tests.
2. Complete Phase 2 to align TypeScript contracts and action helpers with the backend admin API.
3. Complete Phase 3 to deliver visible and editable lifecycle fields.
4. Stop and validate US1 independently with `npm test`, `npm run lint`, `npm run build`, and quickstart checks 1, 2, and 5.

### Incremental Delivery

1. Deliver US1 for lifecycle visibility and edits.
2. Deliver US2 for reversible active/inactive controls.
3. Deliver US3 for permanent delete with confirmation.
4. Run the full polish validation after the selected stories are complete.

### Notes

- Preserve raw key one-time display: raw `api_key` may appear only in the create success dialog and must be cleared when that dialog closes.
- Empty `allowed_models` means unrestricted access; stale model identifiers already present on a key must remain visible rather than being silently dropped.
- Mutation success should call `revalidatePath("/keys")` and client components should refresh immediately plus after short delays to reflect the gateway cache sync window.
- Gateway errors should use the existing error envelope message when available and should be prefixed with the failed action.
