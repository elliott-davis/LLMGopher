# Quickstart: UI Coming Soon Surfaces

## Goal

Replace all eight `Coming soon.` admin pages with scoped UI surfaces that are backed by typed contracts, deterministic fixtures, and focused E2E coverage.

## Prerequisites

- Node dependencies installed in `ui/`.
- Playwright Chromium installed with `npm run test:e2e:install` if the browser is missing.
- No Postgres, Redis, provider credentials, or live gateway are required for mock-backed E2E.

## Implementation Order

1. Add shared UI contract, redaction, and query-state helpers under `ui/src/lib/`.
2. Implement Logs and Audit first because they exercise redaction, URL query filters, empty states, and unavailable states.
3. Implement Teams, Budgets, and Rate Limits using existing fixtures and warning/tripped state assertions.
4. Implement Guardrails with toggle UX gated to mock-backed E2E or reconciled production contracts.
5. Implement Routes with read-only route diagrams and disabled mutation controls until the production route contract exists.
6. Implement Settings with four bounded cards and unavailable/read-only states for backend-dependent controls.
7. Extend visual and accessibility coverage for all eight surfaces.

## Local Verification

From `ui/`:

```bash
npm run lint
npm run test
npm run test:e2e -- tests/e2e/logs.spec.ts tests/e2e/audit.spec.ts tests/e2e/teams.spec.ts tests/e2e/budgets.spec.ts tests/e2e/rate-limits.spec.ts tests/e2e/guardrails.spec.ts tests/e2e/routes.spec.ts tests/e2e/settings.spec.ts
```

For full feature confidence:

```bash
npm run test:e2e
```

From the repository root, run Go checks only if the implementation unexpectedly touches Go code:

```bash
golangci-lint run
go test ./...
```

If `golangci-lint` is unavailable, use `go vet ./...` as the static-analysis fallback.

## Manual Smoke Test

1. Start the UI dev server with the existing E2E/mock setup.
2. Visit `/logs`, apply `5xx` and fallback filters, reload the page, and confirm the selected filter remains in the URL.
3. Open the fallback request inspector and confirm the failed primary and successful fallback stages are visible.
4. Visit `/audit`, apply actor/action/date filters, reload the page, and confirm records remain filtered.
5. Visit `/teams`, `/budgets`, and `/rate-limits`; confirm Research is near cap and exactly one rate-limit rule is tripped.
6. Visit `/guardrails`; toggle `gr_jail` only in mock-backed E2E or a reconciled editable context.
7. Visit `/routes`; confirm strategy diagrams render and production mutation controls are unavailable unless a route API has been reconciled.
8. Visit `/settings`; confirm Gateway Profile, Security, Notifications, and Display cards render and unavailable controls explain why.

## Security Checklist

- No authorization headers, provider credentials, raw API keys, tokens, cookies, or secret-like values render in tables, detail panels, drawers, errors, or settings fields.
- Prompt and response bodies render only as redacted, truncated previews.
- Production writes are not enabled from mock-only contracts.

## Remaining Production API Gaps

The following capabilities are blocked pending real gateway admin API reconciliation:

### Routes (`/routes`)
- **Route mutation** (create, update strategy, delete): the mock returns `501 Not Implemented` for POST/PATCH. Unblocked when `POST /v1/admin/routes` and `PATCH /v1/admin/routes/:id` are reconciled.
- **Live health state**: `health_state` per `RouteTarget` comes from fixtures only. Unblocked when the real gateway returns provider health in the route list response.

### Guardrails (`/guardrails`)
- **CRUD beyond toggle**: only `enabled` toggle is currently wired. Create, delete, and full guardrail edit require expanded admin contract.

### Teams (`/teams`)
- **Member management**: the surface shows member count but member add/remove is out of scope until the team membership admin API ships.
- **Team create/delete**: mock is read-only for team creation.

### Budgets (`/budgets`)
- **Budget create/delete for teams**: PATCH is supported via mock but create/delete requires contract reconciliation.
- **Key-level budget surface**: key-scoped budgets exist in the API key admin surface but are not yet aggregated in the Budgets page.

### Rate Limits (`/rate-limits`)
- **TPM enforcement status**: `tpm` values are stored and displayed but TPM enforcement confirmation is required before the "TPM unavailable" copy is removed.
- **Team-scope rate limits**: team-scoped rules are in the fixture but production enforcement at team scope requires backend confirmation.

### Settings (`/settings`)
- **Gateway Profile, Security, Notifications cards**: all three are `unavailable` until the gateway configuration admin API (`/v1/admin/settings`) is reconciled for these card types.
- **Display preferences persistence**: currently local-only via mock; production persistence requires a user preferences API.

### Logs (`/logs`) and Audit (`/audit`)
- **Pagination**: `next_cursor`-based pagination is stubbed in the mock. Real pagination requires the gateway log API to confirm cursor semantics.
- **Full prompt/response viewing**: currently redacted/truncated preview only. Full viewing is out of scope for the first implementation.
