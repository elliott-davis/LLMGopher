# Surface Spec: Logs

## Page Goal

Give operators a live request log they can filter, scan, and inspect during incident response and routing investigations.

## Source Context

- Current route: `/logs`
- Current state: `Coming soon.`
- Mock contract: `specs/34-ui-e2e-testing-suite/contracts/admin-logs.md`
- Mock fixtures: `ui/tests/fixtures/logs.ts`
- Related specs: `16-provider-fallback`, `17-load-balancing`, `32-ui-usage-audit-dashboard`, `34-ui-e2e-testing-suite`

## Primary Users

- Gateway operator investigating request failures, latency, fallback behavior, or budget/rate-limit denials.
- Administrator validating that routing and provider changes behave as expected.

## Required Capabilities

- Show a paginated newest-first table of gateway request rows.
- Show status, method, path, request id, timestamp, API key id or display label, model, status code, latency, and provider chain summary.
- Filter by status group: all, `2xx`, `4xx`, `5xx`, and fallback.
- Open a request inspector drawer or detail panel from any row.
- In the inspector, show trace, prompt preview, response preview, and headers tabs.
- In the trace tab, show each provider stage with provider id/name, outcome, latency, and failed/successful state.
- For fallback rows, clearly mark the failed primary stage and successful fallback stage.
- Redact authorization headers, provider credentials, raw API keys, and secret-like values before rendering.
- Render prompt and response bodies only as redacted, truncated previews; full raw prompt/response viewing is out of scope for the first implementation.
- Preserve selected filters when opening and closing the inspector.
- Encode active filters in URL query parameters so filtered views survive reloads and can be shared.

## Acceptance Criteria

1. Given 20 seeded log rows, when an operator opens `/logs`, then rows render with mixed 2xx, 4xx, 5xx, and fallback states.
2. Given the operator selects the 5xx filter, when rows render, then every visible row has a 5xx status code.
3. Given the operator selects the fallback filter, when rows render, then each visible row has a multi-stage provider chain with the primary stage failed.
4. Given the operator reloads or shares a filtered Logs URL, when `/logs` renders, then the active filter is restored from query parameters.
5. Given the operator opens `log_fallback`, when the trace tab renders, then the primary stage exposes a failed state and the fallback stage exposes a successful state.
6. Given request detail contains headers, prompts, and responses, when detail tabs render, then authorization and credential-like values are redacted and prompt/response bodies appear only as truncated previews.
7. Given no rows match a filter, when the table renders, then the page shows a useful empty state and a way to clear filters.
8. Given the logs endpoint fails, when `/logs` renders, then the page shows an unavailable state without clearing the active filter controls.

## Data Contract

The page consumes request list rows and request detail from `/v1/admin/logs` and `/v1/admin/logs/{id}`. The first implementation may use the mock contract, but production enablement requires reconciliation with the real gateway contract.

Required row fields:

- `id`
- `request_id`
- `timestamp`
- `method`
- `path`
- `status_code`
- `latency_ms`
- `api_key_id`
- `model`
- `provider_chain`

Required detail-only fields:

- `prompt`
- `response`
- `headers`

## Test Hooks

- `filter-5xx`
- `log-row`
- `log-row-fallback`
- `status`
- `timeline-stage-primary`

## Out Of Scope

- Full-text prompt/response search.
- Full raw prompt/response viewing.
- Exporting log data.
- Real-time streaming updates beyond page refresh or explicit reload.
