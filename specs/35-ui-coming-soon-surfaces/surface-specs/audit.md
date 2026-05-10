# Surface Spec: Audit

## Page Goal

Give administrators a trustworthy, searchable audit history for request activity and administrative changes without direct database access.

## Source Context

- Current route: `/audit`
- Current state: `Coming soon.`
- Mock contract: `specs/34-ui-e2e-testing-suite/contracts/admin-audit.md`
- Mock fixtures: `ui/tests/fixtures/audit.ts`
- Related specs: `06-admin-audit-api`, `08-admin-usage-api`, `32-ui-usage-audit-dashboard`, `34-ui-e2e-testing-suite`

## Primary Users

- Administrator reviewing who or what caused a change or request outcome.
- Operator investigating spend, failures, latency spikes, and policy denials.
- Security reviewer validating that sensitive details are not exposed in admin UI history.

## Required Capabilities

- Show paginated audit records sorted newest-first.
- Filter by actor-like identifier, action/model-like field, and date range.
- Show request id, API key id, model, provider, status, latency, tokens, cost, streaming state, error summary, and timestamp when available.
- Support empty filtered results with clear copy and a reset path.
- Preserve filters across pagination and refresh by encoding active filters in URL query parameters.
- Distinguish successful, failed, unauthorized, and budget-denied records with text labels, not color alone.
- Avoid rendering raw API keys, provider credentials, full prompts, or unredacted secret-like metadata.

## Acceptance Criteria

1. Given seeded audit records, when an administrator opens `/audit`, then at least 10 records are available with request, model, provider, cost, and status context.
2. Given an actor/API-key filter, when it is applied, then every visible record matches the requested actor-like value.
3. Given a date range filter, when it is applied, then every visible record falls within the selected range.
4. Given the administrator reloads or shares a filtered Audit URL, when `/audit` renders, then active filters are restored from query parameters.
5. Given an audit record has an error message, when the row renders, then the error is summarized without exposing credentials or raw request bodies.
6. Given no records match, when the page renders, then the empty state explains the filters and offers a clear-filter action.
7. Given the audit endpoint fails, when `/audit` renders, then the page shows an unavailable state and does not imply the audit log is empty.

## Data Contract

The page consumes `/v1/admin/audit`. The existing backend audit response from `32-ui-usage-audit-dashboard` is the production source of truth. The mock contract may remain narrower as long as it supports the visible UI states.

Required fields when available:

- `id`
- `request_id`
- `api_key_id`
- `model`
- `provider`
- `prompt_tokens`
- `output_tokens`
- `total_tokens`
- `cost_usd`
- `status_code`
- `latency_ms`
- `streaming`
- `error_message`
- `created_at`

## Test Hooks

- Prefer accessible labels for filters.
- Add row-level test ids only when a fixture-specific assertion cannot use role or text queries.

## Out Of Scope

- Editing or deleting audit records.
- Exporting audit logs.
- Compliance retention policy configuration.
