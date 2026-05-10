# Mock Backend Contracts

Each file documents the HTTP shape the mock backend MUST honor for the corresponding endpoint group. The shapes mirror the real LLMGopher gateway — drift between these files and the real gateway is a defect.

| Group | File | Real-gateway source |
|---|---|---|
| Providers | [`admin-providers.md`](./admin-providers.md) | `internal/api/admin_providers*.go` |
| Models | [`admin-models.md`](./admin-models.md) | `internal/api/admin_models*.go` |
| API Keys | [`admin-keys.md`](./admin-keys.md) | `internal/api/admin_keys*.go` |
| Budgets | [`admin-budgets.md`](./admin-budgets.md) | `internal/api/admin_budgets*.go` |
| Teams | [`admin-teams.md`](./admin-teams.md) | `internal/api/admin_teams*.go` (when shipped) |
| Rate Limits | [`admin-rate-limits.md`](./admin-rate-limits.md) | `internal/api/admin_rate_limits*.go` (in flight, spec 33) |
| Guardrails | [`admin-guardrails.md`](./admin-guardrails.md) | spec 26 — not yet shipped |
| Logs | [`admin-logs.md`](./admin-logs.md) | not yet shipped |
| Audit | [`admin-audit.md`](./admin-audit.md) | `internal/api/admin_audit*.go` |

## Conventions

- All requests authenticate with `Authorization: Bearer {LLMGOPHER_UI_ADMIN_API_KEY}`. The mock backend accepts any non-empty token by default and exposes a header `x-mock-require-auth: 1` on the test fixture to flip on strict checking.
- All errors return the OpenAI-compatible envelope `{"error": {"message": "...", "type": "...", "code": "..."}}` with types `authentication_error` | `rate_limit_error` | `invalid_request_error` | `not_found_error` | `internal_error`.
- All timestamps are ISO 8601 UTC.
- All money fields are decimal strings (e.g. `"12.34"`), not numbers, to match the real gateway's serialization.
- Pagination, where present, uses `?limit=&cursor=` and returns `{ data, next_cursor }`.

## Error simulation

A test can request a deterministic error by setting one of:

- Header `x-mock-error: rate_limit_error` (or any of the type values above) — returns 429/4xx with the corresponding envelope.
- Query `?__mock_latency_ms=1500` — sleeps before responding (use sparingly).

These knobs MUST be ignored when `LLMGOPHER_UI_BACKEND_MODE !== "mock"`.
