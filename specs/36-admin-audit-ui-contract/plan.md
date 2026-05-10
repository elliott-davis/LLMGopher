# Implementation Plan: Admin Audit UI Contract

**Branch**: `36-admin-audit-ui-contract` | **Date**: 2026-05-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/36-admin-audit-ui-contract/spec.md`

## Summary

Bring the production `GET /v1/admin/audit` endpoint into vocabulary and behavioral parity with the
admin Audit UI delivered in feature 35: accept `actor`, `action`, `outcome`, date-range, and
pagination filters; redact secret-like substrings from `error_message` at response-shape time;
emit `reference_summary` rows when an audit row references missing entities; and preserve every
existing query parameter and response field for backward compatibility. No `audit_log` schema
changes — all UI-aligned concepts are derived from existing columns. Mock backend
(`ui/tests/mock/handlers/admin-audit.ts`) is updated in lockstep so the UI's E2E tests stay green
against the same shape served by production.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: `net/http` (Go 1.22 ServeMux), `database/sql` + `lib/pq`, goose migrations,
existing `internal/storage` cache and audit query helpers; UI side uses the existing
TypeScript/Playwright mock handler infrastructure (no new deps).
**Storage**: PostgreSQL `audit_log` table (no schema changes); existing async cost worker continues
to write rows.
**Testing**: `go test` with `httptest.NewRecorder()` for handlers, `internal/mocks` for interface
fakes, real Postgres harness for `audit_query_test.go`, table-driven unit tests for the redaction
helper, Playwright E2E for the mock-backend UI run.
**Target Platform**: Linux server (containerized via Docker Compose / kind).
**Project Type**: Web service (Go gateway) with a co-resident React admin UI; this feature is
contract-shape work in the gateway plus a mock-backend adjustment in the UI.
**Performance Goals**: Admin path; no hot-path impact. Endpoint MUST stay under existing admin
latency targets (no new joins). Pagination `count(*)` over filtered set is acceptable.
**Constraints**: OpenAI-compatible error envelope; no DB error leakage (CC-003); additive response
fields only (CC-002); no destructive migrations; existing async cost/audit path unchanged.
**Scale/Scope**: Single endpoint, additive parameters and response fields, three test layers, one
mock-handler update.

## Constitution Check

- **Upstream parity** — PASS. The endpoint remains an LLMGopher-native admin contract (no upstream
  OpenAI equivalent); no divergence in OpenAI-compatible request/response surfaces. Error envelope
  remains `{"error": {"message", "type", "code"}}`.
- **High-throughput runtime** — PASS. Admin path only; no hot-path or streaming changes; no new
  blocking operations on the request proxy path; no new goroutines.
- **Typed contracts** — PASS. New filter inputs and response fields are added as typed Go structs
  in `internal/storage` and `internal/api`; no new `map[string]any` usage.
- **Routing reliability** — N/A. No router, retry, fallback, or rate-limit behavior changes.
- **Multi-tenant spend governance** — PASS. Audit query authorization continues to use the existing
  admin-route protection; `actor` filter is the multi-tenant identity surface.
- **Observability** — PASS. Redaction is strengthened; request/call IDs, structured logs, and
  audit context are preserved; no new external observability calls.
- **API capability UX parity** — PASS. The UI surface (`ui/src/pages/admin/audit/*`) already
  exists from feature 35; this plan brings the production contract into parity with that UI and
  updates the E2E mock handler to match.
- **Security and config** — PASS. Redaction at response time; no credential storage changes; no
  new configuration values; existing precedence rules untouched.
- **Test and lint discipline** — PASS. Three-layer tests (storage, handler, redaction unit) plus a
  mock-handler update. `golangci-lint run` is the preferred enforcement; `go vet ./...` is the
  declared fallback.
- **Linter-first enforcement** — PASS. No new repeatable rule warrants a custom linter; existing
  golangci-lint configuration covers the new code.

No violations. Complexity Tracking section intentionally empty.

## Project Structure

### Documentation (this feature)

```text
specs/36-admin-audit-ui-contract/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── admin-audit.md   # Phase 1 contract for GET /v1/admin/audit
├── checklists/
│   └── requirements.md  # Pre-existing requirements checklist
└── tasks.md             # Phase 2 output (created by /speckit-tasks, not this command)
```

### Source Code (repository root)

```text
internal/
├── api/
│   ├── admin.go                  # Extend HandleGetAuditLog with actor/action/outcome parsing,
│   │                             #   outcome derivation, redaction call, reference_summary emit
│   └── admin_test.go             # Extend; split to admin_audit_test.go if file grows past
│                                 #   maintainability threshold
├── storage/
│   ├── audit_query.go            # Add Actor/Action/Outcome filter inputs + SQL builder
│   ├── audit_query_test.go       # Extend with new filter, ordering, large-offset, empty cases
│   ├── audit_redact.go           # NEW: pure RedactErrorMessage helper
│   └── audit_redact_test.go      # NEW: table-driven redaction tests
└── mocks/                        # Reused as-is; no new mock interfaces required

ui/
└── tests/mock/handlers/
    └── admin-audit.ts            # Update mock to honor actor/action/outcome and emit
                                  #   page/has_more/reference_summary additive fields
```

**Structure Decision**: Single Go web-service repo with a co-resident UI subdirectory. This
feature touches `internal/api` (handler), `internal/storage` (query + redaction helper), and
`ui/tests/mock/handlers` (E2E mock parity). No new packages, no migrations, no new top-level
directories.

## Phase 0: Outline & Research

Complete. See [research.md](./research.md) for the eight resolved decisions covering: UI vocabulary
mapping (R1), action filter semantics (R2), pagination metadata stability (R3), redaction at
response-shape time (R4), missing-reference handling (R5), error envelope and validation (R6),
test surface (R7), and the coordinated mock-backend change (R8). No NEEDS CLARIFICATION remain.

## Phase 1: Design & Contracts

Complete. Outputs:

- [data-model.md](./data-model.md) — Audit Record, Audit Filter, Audit Page Result, Reference
  Summary entities with field-level types, validation rules, and the deterministic outcome
  derivation table.
- [contracts/admin-audit.md](./contracts/admin-audit.md) — Wire contract for `GET /v1/admin/audit`
  including the full query-parameter table, response envelope, validation behavior, and backward-
  compatibility guarantees.
- [quickstart.md](./quickstart.md) — Implementation order, local verification commands, manual
  smoke test (curl examples for both legacy and UI-aligned parameters), and security/compatibility
  checklists.

Agent context: `CLAUDE.md` already references this plan between the SPECKIT markers; no update
required.

### Post-Design Constitution Re-check

Re-evaluated after Phase 1 artifacts: all gates remain PASS. The contract is additive, the
redaction helper is a pure function (testable in isolation), the missing-reference logic does no
new joins, and the mock-handler change keeps the UI parity gate satisfied. No new complexity
introduced beyond what the spec required.

## Complexity Tracking

> No constitution violations to justify.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _none_    | _n/a_      | _n/a_                               |
