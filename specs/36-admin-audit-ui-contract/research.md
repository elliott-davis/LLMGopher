# Phase 0 Research: Admin Audit UI Contract

## R1. Mapping UI vocabulary to existing audit columns

**Decision**: Map UI `actor` to the existing `audit_log.api_key_id` column, alias `api_key_id` as a deprecated-but-supported synonym in the query parser, and prefer `actor` in OpenAPI/UI-facing documentation. Map UI `action` to a small classifier over the existing row (`request:{model}` for request rows; reserved for future admin-action rows). Map UI `outcome` to a derived classifier over `status_code`: `success` for `<400`, `failure` for `≥500`, `unauthorized` for `401`/`403`, `budget_denied` for `429` rows whose `error_message` matches the budget reason, `rate_limited` for other `429` rows, and `client_error` for `4xx` rows that do not match the more specific buckets.

**Rationale**: The UI in feature 35 (and the constitution's API capability UX parity gate) require the production contract to speak the operator's vocabulary. Adding columns to `audit_log` would be a destructive migration and is unnecessary because every UI-facing concept can be derived from existing columns. Keeping `api_key_id` working preserves backward compatibility for existing admin clients (CC-002).

**Alternatives considered**:

- *Add `actor_id`, `action`, `outcome` columns to `audit_log`*: rejected. Requires migration, doubles write cost in the async cost worker, and provides no extra information for current row producers (request flow is the only producer today). Can be revisited if a future spec adds non-request administrative audit rows.
- *Build the mapping in the UI only*: rejected. Each consumer would re-implement the same outcome bucketing, defeating the parity gate and risking divergent labels.
- *Drop `api_key_id` synonym*: rejected. Existing admin scripts and the analytics page already use `api_key_id`; quietly breaking them is outside this feature's scope.

## R2. Action filter semantics for the request-row era

**Decision**: For User Story 1, `action` accepts either an exact action name (`request:{model}`) or a family prefix (`request:` matches all model rows). When omitted, the filter is inert. When the spec requires "exact action requested" (Acceptance 2), the parser distinguishes prefix vs. exact by a trailing `:` or wildcard semantics: `request:` is the family selector, `request:gpt-4o` is the exact selector. The stored column behind the filter is `model` for request-action rows, with a synthesized `request:` prefix added during query builder time only (no schema change).

**Rationale**: The audit table only contains request rows today. Encoding the family in the filter rather than the table keeps the door open for non-request action rows (admin operations) without forcing a wide migration now. The prefix/exact distinction matches operator intuition ("show me everything in this family" vs. "show me this specific call").

**Alternatives considered**:

- *Free-text contains-match*: rejected. Encourages SQL LIKE patterns and is harder to reason about under pagination.
- *Restrict to family-only filter*: rejected. The Acceptance scenarios call out exact-action filtering explicitly.

## R3. Pagination metadata stability

**Decision**: Continue returning `total`, `limit`, and `offset` as the canonical pagination fields. Add `page` (computed as `offset/limit + 1`) and `has_more` (`offset + len(data) < total`) as additive convenience fields the UI can use without recomputation. Sort order remains `created_at DESC, id DESC` so insert-time ties are deterministic.

**Rationale**: `offset`/`limit` is what the UI's `query-state.ts` already serializes from `parseAuditFilter`. Adding `page` and `has_more` lets the UI express "next page" without arithmetic and lets us preserve `total` semantics for spec SC-002 ("at least three pages without duplicate or skipped rows"). The deterministic secondary sort by `id DESC` is critical because two rows can share `created_at` to millisecond precision under high concurrency.

**Alternatives considered**:

- *Cursor-based pagination*: rejected for this feature. Would require a stable, monotonically increasing key and a base64-encoded cursor; out of scope for the UI which already speaks `offset`/`limit` and survives reload via URL query params. Can be added additively in a future feature if scale demands it.
- *Drop `total` for performance*: rejected. The UI relies on `total` to render pagination controls and the spec measurable outcome SC-002 implicitly assumes a stable total. The `count(*)` over the filtered set is acceptable at admin-path latency targets; an index-only count optimization can land later if monitoring shows regressions.

## R4. Redaction at response-shape time vs. storage time

**Decision**: Redaction is applied in `internal/api/admin.go` at the moment the `auditEntryResponse` is built, by a pure helper `internal/storage/audit_redact.go::RedactErrorMessage(string) string`. The function detects bearer tokens (`Bearer ...`), API key prefixes (`sk-...`), authorization headers, cookies, raw long alphanumeric/base64 substrings ≥ 20 chars, and the keywords `key`, `secret`, `token`, `password`, `credential` appearing as standalone words in the error string. Detected substrings are replaced with `[REDACTED]`. The stored row is not mutated.

**Rationale**: Redacting at storage time would be irreversible and would prevent forensic re-analysis if a future investigation needed the raw row (subject to access control). Redacting at response-shape time keeps the database authoritative while still meeting FR-005 and the constitution's Security by Default principle. Putting the helper in `internal/storage/audit_redact.go` (despite the package name) keeps it co-located with the only consumer pair (`audit_query.go` reads, `admin.go` writes the response) and reusable from other admin endpoints.

**Alternatives considered**:

- *Redact on write in the cost worker*: rejected. Causes data loss; if the redaction logic has a bug, the database becomes the bug's permanent record.
- *Redact in middleware*: rejected. Middleware operates on raw bytes; reaching into JSON to find the right field is fragile and slower than a typed-field redaction at the handler.
- *Mirror the UI's TypeScript redaction*: matched in shape (same trigger words, same `[REDACTED]` token) so operators see consistent labels regardless of which side performed the redaction.

## R5. Missing-reference handling

**Decision**: For each row, if the `api_key_id`, `model`, or `provider` value is empty, all-zero, or fails a tiny sanity check (e.g., negative IDs in legacy rows), include a `reference_summary` object in the response indicating the field, the original identifier (or empty string), and a `state` of `missing`, `deleted`, or `unknown`. Do not fail the row, do not fail the query. The original identifier is preserved so an operator can copy it for forensic lookup.

**Rationale**: FR-006 requires the response to "represent missing referenced entities without failing the entire audit query." We can detect the obvious cases at handler time using only the row data. Detecting "deleted" precisely would require a join against `api_keys`, `models`, and `providers` tables, which is reserved for a future enhancement; the `reference_summary.state = "unknown"` covers the cheap case while leaving a clean escalation path.

**Alternatives considered**:

- *Always join all three reference tables*: rejected. Adds three joins to a paginated query that already does a `count(*)`. Too costly for the admin path on large `audit_log` tables.
- *Drop rows that reference missing entities*: rejected. Violates FR-006 explicitly.
- *Hide missing references with empty strings*: rejected. Loses forensic context that the row was incomplete in the first place.

## R6. Error envelope and validation behavior

**Decision**: All validation failures (malformed `from`/`to`, invalid `outcome`, `from > to`, non-positive `limit`, negative `offset`, unknown `actor` syntax) return HTTP 400 with the OpenAI-compatible envelope `{"error": {"message": "...", "type": "invalid_request_error", "code": "..."}}`. The existing handler already follows this pattern; we extend it for `actor`, `action`, and `outcome` with shared validation helpers. Database unavailability returns 503 with `service_unavailable`. Internal errors return 500 with a generic message — the underlying `pq.Error` text MUST NOT leak to clients (CC-003).

**Rationale**: The constitution's principle I (Upstream API & Behavioral Parity) makes the OpenAI-compatible error envelope the only acceptable format. CC-003 explicitly forbids leaking database details. Co-locating new validators with existing ones in `admin.go` keeps the parser pattern consistent.

**Alternatives considered**:

- *Tolerate malformed input by silently dropping it*: rejected. The constitution requires fail-fast validation for admin inputs; silent dropping has caused spec-defying bugs in the past.
- *Return 422 for semantic errors*: rejected. The repo's prior art uses 400 + `invalid_request_error`; consistency outweighs HTTP-status purity here.

## R7. Test surface

**Decision**: Add tests at three layers:

1. *Storage layer*: extend `audit_query_test.go` with cases for actor/action/outcome filters, deterministic ordering across millisecond ties, large-offset pagination, and an empty-result filter.
2. *API layer*: extend `admin_test.go` (and add `admin_audit_test.go` if size grows) for actor/action/outcome parsing, validation failure paths (`invalid_request_error`), redaction integration on `error_message`, missing-reference response shape, and backward-compatible behavior for `api_key_id`/`model`/`provider`/`status` clients.
3. *Redaction unit*: a focused `audit_redact_test.go` for the pure helper covering bearer tokens, sk- prefixes, authorization headers, cookies, base64-like long substrings, and case-insensitive keyword matches.

**Rationale**: This matches the constitution's Test Discipline principle and the `Test Coverage Target` rule in CLAUDE.md (80%). The three-layer split keeps each test file focused and lets the coverage tool report changes per concern.

**Alternatives considered**:

- *End-to-end tests against a live Postgres in CI*: useful but out of scope for this feature; the existing `audit_query_test.go` already exercises a real Postgres in the test environment via the same setup.
- *Property-based tests for redaction*: not justified for a feature this narrow; table-driven tests cover the named cases in FR-005.

## R8. Coordinated changes with the UI mock backend

**Decision**: Update `ui/tests/mock/handlers/admin-audit.ts` so the mock contract returns the same UI-aligned shape (accepts `actor`/`action`/`outcome` query params, returns `total`, `limit`, `offset`, `page`, `has_more`, and a `reference_summary` field where applicable). This keeps the UI's E2E tests green even though the UI was already written against this shape; the mock had been a slimmer subset. Production tests against the real gateway must independently pass — the mock is for E2E only.

**Rationale**: The UI uses the mock backend as the deterministic source for E2E tests. Diverging the mock from the production contract would silently mask integration bugs, defeating the parity gate.

**Alternatives considered**:

- *Leave the mock alone and add a "production-only" flag in tests*: rejected. Two contracts means two redaction implementations and a maintenance treadmill.
- *Implement the changes only in the mock first, defer Go work*: rejected. The whole feature exists to align production with what the UI already does; doing only mock work is a no-op against the spec.
