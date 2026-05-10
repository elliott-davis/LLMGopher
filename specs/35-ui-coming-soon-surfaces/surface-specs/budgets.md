# Surface Spec: Budgets

## Page Goal

Let administrators monitor and adjust spend controls across team and key scopes, with clear warning and hard-cap states.

## Source Context

- Current route: `/budgets`
- Current state: `Coming soon.`
- Mock contract: `specs/34-ui-e2e-testing-suite/contracts/admin-budgets.md`
- Mock fixtures: `ui/tests/fixtures/budgets.ts` and key budgets in `ui/tests/fixtures/keys.ts`
- Related specs: `07-admin-budget-api`, `10-budget-lifecycle`, `31-ui-key-budgets`, `34-ui-e2e-testing-suite`

## Primary Users

- Administrator managing spend limits for teams and API keys.
- Operator identifying near-cap and over-cap tenants during request failures.

## Required Capabilities

- Show budget policies grouped by scope, starting with team budgets and linking to key-level budget context where available.
- Show limit, usage, duration, alert threshold, utilization percent, and warning/hard-cap state.
- Mark budgets at or above alert threshold with text and visual treatment.
- Support editing limit, duration, and alert threshold only when a reconciled production mutation contract or mock-backed E2E fixture supports mutation.
- Preserve usage values when budget policy settings change.
- Confirm successful saves and show clear gateway error feedback on failure.
- Separate "no budgets configured" from "budget service unavailable."

## Acceptance Criteria

1. Given Research has monthly limit `1000.00`, usage `860.00`, and threshold `0.85`, when `/budgets` renders, then Research shows a near-cap warning.
2. Given Platform has monthly limit `2000.00`, usage `800.00`, and threshold `0.85`, when `/budgets` renders, then Platform does not show a warning.
3. Given a reconciled production mutation contract or mock-backed E2E fixture supports team budget edits, when an administrator changes a team's limit, duration, or alert threshold and the save succeeds, then the updated policy renders while current usage remains intact.
4. Given a budget save fails in an editable context, when the administrator submits budget changes, then the form preserves entered values and displays the failure reason.
5. Given a key is over hard cap, when a request is made with that key in the mock contract test, then the gateway responds with a budget-exceeded reason.
6. Given no budgets exist, when the page renders, then the empty state explains how budgets are created or linked.

## Data Contract

The page consumes `/v1/admin/budgets` and mutates `/v1/admin/budgets/{scope}/{scope_id}` for supported scopes. Key-level budget capabilities remain aligned with existing key budget UI and APIs.

Required fields:

- `scope`
- `scope_id`
- `limit_usd`
- `usage_usd`
- `duration`
- `alert_threshold`

## Test Hooks

- `team-research-warn`
- Prefer accessible labels for edit controls.

## Out Of Scope

- Real-time budget metering beyond displayed usage refresh.
- Budget forecasts.
- Team membership management.
