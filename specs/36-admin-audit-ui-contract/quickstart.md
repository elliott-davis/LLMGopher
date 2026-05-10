# Quickstart: Admin Audit UI Contract

## Goal

Bring the production `GET /v1/admin/audit` endpoint into parity with the language and behavior the admin Audit UI already exposes (feature 35). Filtering by actor, action, date range, and outcome MUST work end-to-end against the real gateway, not only against the mock backend.

## Prerequisites

- Go 1.22+ with the LLMGopher repository checked out.
- A local PostgreSQL with `make dev` already producing a populated `audit_log` table, or the existing test harness used by `internal/storage/audit_query_test.go`.
- `golangci-lint` available locally if possible (`brew install golangci-lint`); otherwise `go vet ./...` is the fallback.
- `make dev` for end-to-end verification through the UI.

## Implementation Order

1. Add the redaction helper at `internal/storage/audit_redact.go` and its unit tests at `internal/storage/audit_redact_test.go`.
2. Extend `internal/storage/audit_query.go` with the `Actor`, `Action`, and `Outcome` filter inputs and update the SQL builder to handle the new predicates without breaking the existing ones. Extend `internal/storage/audit_query_test.go` to cover them, including ordering and large-offset pagination.
3. Extend `internal/api/admin.go::HandleGetAuditLog` with the new query parameter parsing and the `outcome` derivation. Apply redaction to `error_message` and emit `reference_summary` when a row's references are missing/unknown. Extend `internal/api/admin_test.go` (or split into `admin_audit_test.go`) to cover backward compatibility, the new filters, validation failures, redaction, and missing-reference rows.
4. Update `ui/tests/mock/handlers/admin-audit.ts` so the mock contract honors the same parameters and emits the same response envelope. No UI source changes are required.
5. Verify linting, unit, integration, and a focused E2E run.

## Local Verification

From the repository root:

```bash
golangci-lint run ./... || go vet ./...
go test ./internal/api/... ./internal/storage/... -count=1
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -20
```

Confirm coverage on `internal/api` and `internal/storage` is at or above 80% on changed files.

For end-to-end UI confidence:

```bash
make dev
( cd ui && npm run test:e2e -- tests/e2e/audit.spec.ts )
```

The Audit page should render filtered, redacted, newest-first rows whether the UI is pointed at the mock backend (default for E2E) or the real gateway.

## Manual Smoke Test

1. Start the gateway with `make run`.
2. Curl the endpoint with the existing parameter names and confirm the response is unchanged in shape (additive fields permitted):

   ```bash
   curl -s 'http://localhost:8080/v1/admin/audit?api_key_id=key_checkout_service&model=gpt-4o&status=success&limit=5'
   ```

3. Curl the endpoint with the new UI-aligned parameters:

   ```bash
   curl -s 'http://localhost:8080/v1/admin/audit?actor=key_checkout_service&action=request:gpt-4o&outcome=success&limit=5'
   ```

4. Confirm both calls return the same `data` rows (modulo additive fields), the same `total`, and matching `page`/`has_more` values.
5. Force a redaction case by inserting an audit row whose `error_message` contains `Bearer test-token-123` and confirm the response shows `[REDACTED]` in place of the bearer value while preserving surrounding text.
6. Force a missing-reference case by inserting an audit row whose `api_key_id` is empty and confirm the response includes a `reference_summary` entry with `field = "actor_id"`, `state = "missing"`.

## Security Checklist

- No authorization headers, provider credentials, raw API keys, tokens, cookies, or secret-like substrings appear in `error_message` after redaction.
- Database error text MUST NOT leak through the OpenAI-compatible error envelope.
- Existing admin authentication and authorization rules remain unchanged.

## Compatibility Checklist

- Existing query parameters (`api_key_id`, `model`, `provider`, `status`, `from`, `to`, `limit`, `offset`) continue to return functionally identical row sets.
- Existing response fields (`id`, `request_id`, `api_key_id`, `model`, `provider`, token counts, `cost_usd`, `status_code`, `latency_ms`, `streaming`, `error_message`, `created_at`, `total`, `limit`, `offset`) remain present and typed.
- New response fields (`actor_id`, `action`, `outcome`, `page`, `has_more`, `reference_summary`) are additive.
- `actor` and `api_key_id` are mutually exclusive on the request side; specifying both returns 400.
- `outcome` and `status` may both be specified; `outcome` wins.

## Out of Scope

- `GET /v1/admin/audit/:id` (single-row detail) is not in this feature.
- Joining against `api_keys`, `models`, and `providers` to detect `state = "deleted"` precisely is reserved for a follow-up; until then only `missing` and `unknown` are emitted.
- Cursor-based pagination is reserved for a future feature.
- Schema changes to `audit_log` are explicitly out of scope; this feature is a contract-shape feature.
- Non-request action rows (e.g., admin operations) are reserved for a future feature; the `action` field is forward-compatible with future prefixes such as `admin:`.
