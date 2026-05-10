# Surface Spec: Guardrails

## Page Goal

Let administrators view and manage gateway guardrail policies, starting with enabled/disabled state and expanding only where backend guardrail contracts exist.

## Source Context

- Current route: `/guardrails`
- Current state: `Coming soon.`
- Mock contract: `specs/34-ui-e2e-testing-suite/contracts/admin-guardrails.md`
- Mock fixtures: `ui/tests/fixtures/guardrails.ts`
- Related specs: `26-guardrail-integrations`, `34-ui-e2e-testing-suite`

## Primary Users

- Administrator responsible for prompt, response, PII, jailbreak, and secret-scanning controls.
- Operator confirming whether safety controls are active before or during an incident.

## Required Capabilities

- List guardrails with display name, enabled state, policy category when available, and concise description when available.
- Toggle enabled state for an existing guardrail only when a reconciled production mutation contract or mock-backed E2E fixture supports the mutation.
- Persist toggle state across reloads after successful save in editable contexts.
- Show a saving state while a toggle mutation is in progress in editable contexts.
- Restore or preserve state clearly when a save fails in editable contexts.
- Distinguish unavailable backend support from an empty guardrail list.
- Avoid exposing guardrail provider credentials, raw detector prompts, or sensitive match payloads.

## Acceptance Criteria

1. Given seeded guardrails `gr_jail`, `gr_pii`, and `gr_secrets`, when an administrator opens `/guardrails`, then all three render with their current enabled state.
2. Given `gr_jail` is disabled in a reconciled production mutation contract or mock-backed E2E fixture, when the administrator toggles it on, then it renders as enabled after save and remains enabled after reload.
3. Given a toggle request fails in an editable context, when the administrator toggles a guardrail, then the page displays the failure and does not leave the final state ambiguous.
4. Given no guardrails exist, when the page renders, then the empty state explains that no policies are configured and points to supported integrations.
5. Given guardrail details include sensitive matching context, when the page renders, then sensitive payloads are not displayed.

## Data Contract

The page consumes `/v1/admin/guardrails` and mutates `/v1/admin/guardrails/{id}`. The mock contract covers `id`, `display_name`, and `enabled`. Production implementation must reconcile with `26-guardrail-integrations` before adding richer configuration.

Required fields:

- `id`
- `display_name`
- `enabled`

Optional fields for future enhancement:

- category
- description
- provider/integration label
- last updated timestamp
- last match summary

## Test Hooks

- `toggle-{id}`, including `toggle-gr_jail`

## Out Of Scope

- Creating new guardrail policies before the backend contract exists.
- Editing detector rules, thresholds, or provider credentials.
- Displaying raw blocked prompt/response content.
