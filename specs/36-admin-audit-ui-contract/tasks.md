---

description: "Task list for feature 36 — Admin Audit UI Contract"
---

# Tasks: Admin Audit UI Contract

**Input**: Design documents from `/specs/36-admin-audit-ui-contract/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/admin-audit.md, quickstart.md

**Tests**: Required. Spec FR-009 calls out success and failure tests for filtering, pagination,
invalid input, redaction, and missing references. Constitution Principle VIII requires unit,
integration, contract, failure-path, and (where applicable) UI tests written before or alongside
implementation. `golangci-lint run` is required when available.

**Organization**: Tasks are grouped by user story so each story can be implemented and validated
independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no incomplete dependencies)
- **[Story]**: User story tag (US1, US2)
- All paths are absolute or repository-relative

## Path Conventions

- Go gateway code: `internal/api/`, `internal/storage/`, `internal/mocks/`
- UI mock backend: `ui/tests/mock/handlers/`
- Spec/contract docs: `specs/36-admin-audit-ui-contract/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm tooling and baseline state before changing the audit contract.

- [x] T001 Verify local toolchain by running `go vet ./...` and (if available) `golangci-lint run` from the repository root; record any pre-existing warnings so new ones are distinguishable.
- [x] T002 [P] Confirm `make dev` brings up gateway + Postgres + Redis + UI and that `GET /v1/admin/audit?limit=1` returns the current legacy shape; capture the baseline JSON for later regression comparison.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Land shared infrastructure that both user stories depend on. Both US1 (filtering) and US2 (safe display) consume the same response shape, the same outcome derivation, and the same redaction helper, so these MUST be in place before either story is wired through `HandleGetAuditLog`.

**⚠️ CRITICAL**: No user-story phase can be merged until this phase is complete.

- [x] T003 [P] Confirm upstream/native parity decision in `specs/36-admin-audit-ui-contract/research.md` (R1, R6) — the endpoint is LLMGopher-native; OpenAI-compatible error envelope is mandatory. No code change; this is the divergence-notes confirmation required by the plan's Constitution Check.
- [x] T004 [P] Add typed request-filter and response structs to `internal/storage/audit_query.go`: extend `AuditQueryFilter` with `Actor`, `Action`, and `Outcome` fields per `data-model.md`; add `ActionExact bool` (or equivalent) so the SQL builder distinguishes `request:` family from `request:{model}` exact selectors per research R2.
- [x] T005 [P] Add the typed response shapes (`AuditRecordResponse`, `AuditPageResult`, `ReferenceSummary`, and the `outcome` enum constants) to `internal/api/admin.go` (or a new `internal/api/admin_audit_types.go` if `admin.go` grows past maintainability) per `data-model.md`.
- [x] T006 Create the pure redaction helper at `internal/storage/audit_redact.go` exporting `RedactErrorMessage(string) string`. Implementation MUST detect bearer tokens, `sk-` API key prefixes, authorization headers, cookies, base64-like substrings ≥20 chars, and the standalone keywords `key`/`secret`/`token`/`password`/`credential`, replacing matches with `[REDACTED]` while preserving surrounding human-readable text (research R4).
- [x] T007 Add the deterministic outcome-derivation function (input: `status_code`, `error_message`; output: one of `success`, `client_error`, `unauthorized`, `rate_limited`, `budget_denied`, `failure`) in `internal/api/admin.go`, matching the table in `data-model.md` §Outcome Derivation Rules. Function MUST be pure and exported within-package for direct unit testing.
- [x] T008 Add the missing-reference detector in `internal/api/admin.go` that emits `[]ReferenceSummary` for empty/zero/negative `actor_id`, `model`, or `provider` values per research R5; deleted-state lookups are explicitly out of scope (only `missing` and `unknown` emitted).

**Checkpoint**: Foundation ready — both user stories can now be wired through the handler in parallel.

---

## Phase 3: User Story 1 — Filter Audit History for UI Investigations (Priority: P1) 🎯 MVP

**Goal**: Operators can filter audit history with the UI's vocabulary (`actor`, `action`, `outcome`, `from`, `to`, pagination) against the production endpoint and receive newest-first, deterministically-ordered, paginated results with stable `total`/`page`/`has_more` metadata.

**Independent Test**: Seed `audit_log` with rows spanning multiple actors, actions, outcomes, and timestamps; issue `GET /v1/admin/audit?actor=...&action=request:gpt-4o&outcome=success&from=...&to=...&limit=20&offset=0`; verify every returned row matches the filter, ordering is `created_at DESC, id DESC`, and paginating across at least three pages produces disjoint, contiguous slices with no duplicates and no gaps (covers SC-001, SC-002).

### Tests for User Story 1

> Write these tests FIRST and confirm they fail before T015–T019 implementation.

- [x] T009 [P] [US1] Extend `internal/storage/audit_query_test.go` with cases for `Actor`, `Action` (family prefix `request:` and exact `request:{model}`), and each `Outcome` value mapping to the predicate matrix from `data-model.md`.
- [x] T010 [P] [US1] Extend `internal/storage/audit_query_test.go` with a deterministic-ordering case: insert rows that share `created_at` to millisecond precision and assert `id DESC` is the tiebreaker.
- [x] T011 [P] [US1] Extend `internal/storage/audit_query_test.go` with a large-offset pagination case (≥3 pages of `limit=20`) asserting disjoint, contiguous coverage of the filtered set; covers SC-002.
- [x] T012 [P] [US1] Add handler-layer tests in `internal/api/admin_test.go` (or new `internal/api/admin_audit_test.go`) for parsing each new query parameter, including the `actor`/`api_key_id` mutual-exclusion 400 (`code: "ambiguous_actor"`) and the `outcome`-wins-over-`status` precedence rule.
- [x] T013 [P] [US1] Add handler-layer validation-failure tests in `internal/api/admin_test.go` for malformed `from`/`to`, `from > to`, unknown `outcome`, non-positive `limit`, and negative `offset`, asserting `invalid_request_error` with the OpenAI-compatible envelope per research R6.
- [x] T014 [P] [US1] Add a handler-layer pagination test in `internal/api/admin_test.go` asserting `total`, `limit`, `offset`, `page = offset/limit + 1`, and `has_more = offset + len(data) < total` are all populated and consistent across pages.

### Implementation for User Story 1

- [x] T015 [US1] Extend the SQL builder in `internal/storage/audit_query.go` to translate the new filter fields into predicates: `Actor` → `api_key_id = $n`; `Action` (family) → `model IS NOT NULL` plus future-action guard; `Action` (exact `request:{model}`) → `model = $n`; each `Outcome` value → the deterministic predicate set from `data-model.md` (`budget_denied` uses `status_code = 429 AND error_message ILIKE '%budget%'`).
- [x] T016 [US1] Update the existing query in `internal/storage/audit_query.go` to enforce `ORDER BY created_at DESC, id DESC` (add the `id DESC` tiebreaker if absent) and ensure the `count(*)` runs against the same `WHERE` clause as the data query.
- [x] T017 [US1] Extend `HandleGetAuditLog` in `internal/api/admin.go` to parse `actor`, `action`, `outcome`, and pagination params, applying the precedence rules (`actor` wins over `api_key_id`; `outcome` wins over `status`) and the validation behavior covered by T012/T013.
- [x] T018 [US1] Build the response envelope in `internal/api/admin.go` to populate `data`, `total`, `limit`, `offset`, `page`, and `has_more` per `data-model.md` §Audit Page Result; for each row, populate `actor_id` (mirrored from `api_key_id` for request rows), `action` (`request:{model}`), and `outcome` (from T007).
- [x] T019 [US1] Update the UI mock backend at `ui/tests/mock/handlers/admin-audit.ts` so it accepts `actor`/`action`/`outcome` query params and emits `page`/`has_more` (and any other additive fields exercised by US1) so the UI's E2E tests stay green against the same shape as production (research R8).

**Checkpoint**: US1 is independently testable — `make test` passes the storage and handler tests, and `( cd ui && npm run test:e2e -- tests/e2e/audit.spec.ts )` passes against the mock backend.

---

## Phase 4: User Story 2 — Preserve Safe Audit Detail Display (Priority: P2)

**Goal**: Audit rows surfaced through the admin contract carry investigation context (outcome label, reference summary) without leaking secrets in `error_message`, and the response remains usable when an audit row references a missing/unknown actor, model, or provider.

**Independent Test**: Insert audit rows whose `error_message` contains `Bearer ...`, `sk-...`, raw long alphanumerics, and the words `key`/`secret`/`token`; insert one row whose `api_key_id` is empty; request the rows through `GET /v1/admin/audit`; verify all sensitive substrings are replaced with `[REDACTED]` while surrounding text remains, every outcome is distinguishable by text (not color-only), and the empty-`api_key_id` row carries a `reference_summary` entry with `field = "actor_id"`, `state = "missing"` (covers SC-003 and FR-005/FR-006).

### Tests for User Story 2

> Write these tests FIRST and confirm they fail before T023–T025 implementation.

- [x] T020 [P] [US2] Add `internal/storage/audit_redact_test.go` with table-driven cases for: `Bearer <token>`, `Authorization: Bearer ...`, `sk-...` prefixes (mixed case), `Cookie: sid=...`, base64-like ≥20-char substrings, and standalone keywords `key`/`secret`/`token`/`password`/`credential` (case-insensitive); each case asserts the secret substring is replaced with `[REDACTED]` and the surrounding human-readable text (including the keyword `budget` for outcome derivation) is preserved.
- [x] T021 [P] [US2] Add a handler-layer redaction integration test in `internal/api/admin_test.go` (or `internal/api/admin_audit_test.go`) that seeds an audit row with secret-bearing `error_message`, calls `HandleGetAuditLog`, and asserts the response body contains `[REDACTED]` and does not contain the original secret.
- [x] T022 [P] [US2] Add a handler-layer missing-reference test in `internal/api/admin_test.go` covering empty `api_key_id`, empty `model`, and empty `provider`; assert the row is still returned, the `reference_summary` entry has the correct `field`/`original_id`/`state` per `data-model.md`, and the query does not fail (FR-006).

### Implementation for User Story 2

- [x] T023 [US2] Wire `RedactErrorMessage` (T006) into the response builder in `internal/api/admin.go` so every `error_message` is redacted before serialization. Order matters: redaction MUST run BEFORE outcome derivation (T007) reads the message, since redaction preserves the keyword `budget` per research R4.
- [x] T024 [US2] Wire the missing-reference detector (T008) into the response builder so each row's `reference_summary` is populated when applicable; the field MUST be omitted (not `null`) when all references are intact, per `data-model.md`.
- [x] T025 [US2] Update `ui/tests/mock/handlers/admin-audit.ts` to surface `reference_summary` and to leave any redaction triggers in the seeded fixtures alone (the mock does not need to re-implement redaction; it only needs to honor the response shape so UI rendering tests cover both the redacted-text and missing-reference paths).

**Checkpoint**: US2 is independently testable — redaction unit tests pass, the handler-layer redaction and missing-reference tests pass, and the UI mock-backed E2E run continues to render redacted text and reference-summary indicators without color-only state.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Validation, observability, compatibility checks, and lint hygiene that span both user stories.

- [x] T026 [P] Run `go test ./internal/api/... ./internal/storage/... -count=1` and `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -20`; confirm changed files in `internal/api` and `internal/storage` are at or above 80% coverage (Constitution Principle VIII; CLAUDE.md test coverage target).
- [x] T027 [P] Run `golangci-lint run ./...` (or `go vet ./...` as the documented fallback) and resolve any new warnings introduced by this feature; if a repeatable rule emerges during review, prefer adding a golangci-lint configuration over an LLM-only review note (Constitution Principle VIII; plan's Linter-first enforcement gate).
- [x] T028 [P] Run the manual smoke test from `specs/36-admin-audit-ui-contract/quickstart.md` §Manual Smoke Test, including the legacy-parameter curl, the UI-aligned-parameter curl, the redaction case, and the missing-reference case; capture transcripts in the PR description for compatibility review.
- [x] T029 [P] Run `( cd ui && npm run test:e2e -- tests/e2e/audit.spec.ts )` against the updated mock backend to confirm the UI's audit page still renders filtered, redacted, newest-first rows with reference-summary annotations.
- [x] T030 Review `specs/36-admin-audit-ui-contract/contracts/admin-audit.md` against the implemented response and update if any field name, type, or precedence rule diverged during implementation; the contract document is the source of truth for downstream admin clients (CC-002).
- [x] T031 Confirm async cost/audit path is unchanged: `internal/proxy/cost_worker.go` still writes raw (un-redacted) rows to `audit_log`; redaction is response-shape-only per research R4 and the constitution's async accounting rule.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup. Blocks both user stories. T004/T005 are independent typed-shape additions; T006 is independent; T007 and T008 are co-located in `admin.go` so should be done sequentially or carefully merged.
- **User Story 1 (Phase 3)**: Depends on Foundational. Storage tests (T009–T011) gate T015–T016. Handler tests (T012–T014) gate T017–T018. T019 (mock backend) is independent of the Go work and can run in parallel once shape is agreed.
- **User Story 2 (Phase 4)**: Depends on Foundational (specifically T006 redaction helper and T008 missing-reference detector). Independent of US1 implementation but shares the handler file (`admin.go`), so wiring tasks T023–T024 should land after US1's T017–T018 to minimize merge conflicts.
- **Polish (Phase 5)**: Depends on both US1 and US2 being complete.

### Within Each User Story

- Tests MUST be written and FAIL before implementation tasks.
- Storage layer changes precede handler-layer changes.
- Handler-layer changes precede mock-backend updates.

### Parallel Opportunities

- T004, T005, T006 (foundational typed shapes + redaction helper) can run in parallel — different files.
- All US1 test tasks T009–T014 can run in parallel — different test cases, different files.
- All US2 test tasks T020–T022 can run in parallel — different files.
- T019 (UI mock for US1) and T025 (UI mock for US2) touch the same file; sequence them.
- T026–T029 in Polish can run in parallel.

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all US1 test scaffolds in parallel:
Task: "Storage filter tests for actor/action/outcome in internal/storage/audit_query_test.go"
Task: "Storage deterministic-ordering test in internal/storage/audit_query_test.go"
Task: "Storage large-offset pagination test in internal/storage/audit_query_test.go"
Task: "Handler param-parsing tests in internal/api/admin_test.go"
Task: "Handler validation-failure tests in internal/api/admin_test.go"
Task: "Handler pagination metadata test in internal/api/admin_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: SC-001 and SC-002 met against seeded data; UI E2E green against the updated mock.
5. Demo or deploy if ready.

### Incremental Delivery

1. Setup + Foundational → infrastructure ready.
2. US1 → demo filtered/paginated audit history end-to-end.
3. US2 → demo redaction + reference-summary safety.
4. Polish → coverage, lint, smoke, contract sign-off.

### Parallel Team Strategy

- One engineer drives Foundational T006 (redaction helper, isolated file).
- A second engineer drives Foundational T004 + US1 storage and handler work.
- A third engineer drives Foundational T008 + US2 redaction wiring and tests.
- Mock-backend updates (T019, T025) coordinated by whoever owns the UI side; they are small but share `ui/tests/mock/handlers/admin-audit.ts`.

---

## Notes

- [P] tasks = different files, no incomplete dependencies.
- Spec FR-009 mandates failure-path tests for filtering, pagination, invalid input, redaction, and missing references; T012/T013/T020/T021/T022 cover these.
- Async cost/audit path is intentionally untouched (Constitution Principle II; CLAUDE.md async invariant); T031 verifies.
- Commit after each task or logical group; auto-commit hooks are enabled in `.specify/extensions.yml`.
- Stop at the US1 checkpoint to validate the MVP independently before starting US2.
