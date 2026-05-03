# Quickstart: UI Usage and Audit Dashboard

## Prerequisites

- Start the development stack with Postgres, Redis, gateway, and UI:

```bash
make dev
```

- Configure the UI service with an admin gateway token:

```bash
LLMGOPHER_UI_ADMIN_API_KEY=sk-test-key-1
```

The token is required because `/v1/admin/usage`, `/v1/admin/usage/daily`, and `/v1/admin/audit` are admin-authenticated endpoints.

## Manual Verification

1. Open the admin UI.
2. Confirm the sidebar and dashboard include `Usage & Audit`.
3. Open `/usage`.
4. Select `model`, `provider`, and `api_key` grouping and verify the usage summary updates.
5. Set a multi-day time window and verify daily trend rows render one day at a time without crashing on missing days.
6. Filter audit records by `status=error`, model, provider, API key, and time range.
7. Use audit pagination and confirm filters are preserved.
8. Remove or invalidate `LLMGOPHER_UI_ADMIN_API_KEY` and confirm the page shows an unavailable/auth guidance state without exposing secrets.

## Focused Test Commands

Run UI tests:

```bash
cd ui && npm test -- --run
```

Run UI build/lint checks:

```bash
cd ui && npm run lint && npm run build
```

Run backend checks only if Go files are changed:

```bash
go test ./...
golangci-lint run
```

If `golangci-lint` is unavailable, use:

```bash
go vet ./...
```

## Expected Implementation Checks

- `ui/src/lib/analytics.test.ts` covers query construction, admin token behavior, error mapping, and response parsing.
- `ui/src/app/(dashboard)/usage/page.test.tsx` covers non-empty, empty, unavailable, invalid-filter, and pagination states.
- Existing key/model/provider UI tests continue to pass after navigation/dashboard changes.

## Implementation Notes

- Remaining analytics and audit capability gaps: **none** for this feature scope. `/v1/admin/usage`, `/v1/admin/usage/daily`, and `/v1/admin/audit` are now surfaced in `/usage`.
- Backend Go handlers remained read-only during this implementation; Go test/lint reruns were skipped intentionally per the task plan.
