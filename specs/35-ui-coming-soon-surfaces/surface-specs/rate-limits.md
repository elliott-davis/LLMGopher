# Surface Spec: Rate Limits

## Page Goal

Give administrators a central view of throttling rules across model, key, and team scopes, including active or recently tripped rules.

## Source Context

- Current route: `/rate-limits`
- Current state: `Coming soon.`
- Mock contract: `specs/34-ui-e2e-testing-suite/contracts/admin-rate-limits.md`
- Mock fixtures: `ui/tests/fixtures/rate-limits.ts`
- Related specs: `09-per-model-rate-limits`, `33-ui-model-rate-limits`, `23-teams-organizations`, `34-ui-e2e-testing-suite`

## Primary Users

- Administrator managing request and token throttling policies.
- Operator investigating rate-limit denials and noisy tenants/models.

## Required Capabilities

- List rate-limit rules by scope, scope identifier, request-per-second limit, token-per-minute limit when available, and tripped state.
- Clearly distinguish model-scoped, key-scoped, and team-scoped rules.
- Show a "tripped" indicator for rules currently or recently reported as tripped.
- Create, edit, and delete rules only when a reconciled production mutation contract or mock-backed E2E fixture supports those mutations.
- Validate non-negative numeric limit fields and require at least one meaningful limit.
- Explain the relationship between this central page and model-level controls already exposed on the Models page.
- Avoid implying token-per-minute enforcement exists in production until the backend contract confirms it.

## Acceptance Criteria

1. Given seeded rules `rl_chat_default`, `rl_key_checkout`, and `rl_tripped`, when `/rate-limits` renders, then all three rules show scope, scope id, and RPS value.
2. Given exactly one seeded rule has `tripped: true`, when the page renders, then exactly one row shows a tripped indicator.
3. Given a rule has TPM configured, when the row renders, then TPM is visible separately from RPS.
4. Given an administrator submits a negative RPS or TPM value, when the form validates, then the page rejects the value with clear feedback.
5. Given a reconciled production mutation contract or mock-backed E2E fixture supports deleting a rate-limit rule, when a delete request succeeds and the list refreshes, then the deleted rule no longer appears.
6. Given the rate-limit endpoint is unavailable, when the page renders, then the page shows an unavailable state and keeps model-level rate-limit guidance visible.

## Data Contract

The page consumes and may mutate `/v1/admin/rate-limits`. The mock contract supports standard CRUD, but production implementation must reconcile with actual backend support from the rate-limit features before enabling writes.

Required fields:

- `id`
- `scope`
- `scope_id`
- `rps`
- `tripped`

Optional fields:

- `tpm`

## Test Hooks

- Tripped row indicator should be discoverable by accessible text.
- Add rule-specific test ids only if fixture-specific assertions cannot use role or text queries.

## Out Of Scope

- Per-endpoint or per-HTTP-method throttling.
- Historical rate-limit analytics.
- Automatic remediation or cooldown tuning.
