# Surface Spec: Routes

## Page Goal

Let administrators inspect and configure how gateway traffic is routed across providers and models, including fallback, weighted, latency-aware, and single-provider strategies.

## Source Context

- Current route: `/routes`
- Current state: `Coming soon.`
- Mock/UI test expectation: `ui/tests/e2e/routes.spec.ts`
- Related specs: `16-provider-fallback`, `17-load-balancing`, `18-context-window-fallback`, `15-provider-retry`, `34-ui-e2e-testing-suite`

## Primary Users

- Administrator configuring model-to-provider routing.
- Operator validating that provider health and fallback behavior are reflected before changing traffic.

## Required Capabilities

- List configured routes by route/model alias, strategy, primary provider, fallback providers, enabled state, and last-known health context.
- Support strategy views for single-provider, fallback, weighted, and latency-aware routing.
- Show a route diagram that visually and textually represents the selected strategy.
- For fallback strategy, show primary and secondary paths and distinguish fallback paths with a dashed or otherwise accessible secondary state.
- For weighted strategy, expose adjustable provider weights and reflect the changed weights in the diagram.
- For latency-aware strategy, show provider health or latency signals used for operator decision-making.
- When a reconciled production mutation contract or mock-backed E2E fixture supports edits, validate route changes before save so invalid provider mixes, missing primary providers, and zero-total weights are rejected.
- Preserve existing provider/model configuration behavior and avoid changing runtime routing until an administrator saves through a reconciled production mutation contract.

## Acceptance Criteria

1. Given configured providers and models, when an administrator opens `/routes`, then each route shows strategy, providers, and current policy summary.
2. Given a route supports fallback, when the administrator selects the fallback variant, then the diagram shows one primary path and at least one clearly differentiated secondary fallback path.
3. Given a route supports weighted routing in a reconciled production mutation contract or mock-backed E2E fixture, when a provider weight changes, then the visual path weight and summary values update before save.
4. Given a route change is invalid in an editable route form, when the administrator attempts to save, then the page prevents the change and explains the specific issue.
5. Given route save fails in a reconciled production mutation contract or mock-backed E2E fixture, when the gateway returns an error, then the page preserves edits and displays the failure reason.
6. Given no route data is available, when `/routes` renders, then the page shows an empty state that points to provider/model setup as the prerequisite.

## Data Contract

No dedicated mock route contract exists yet. First implementation must define a typed admin route payload before enabling writes. At minimum, the UI needs:

- route id or alias
- model alias
- strategy
- enabled state
- provider targets
- provider order for fallback
- provider weights for weighted routing
- health/latency summary for operator display

## Test Hooks

- `route-{slug}`
- Strategy diagram selector for path assertions
- Weight controls with accessible labels

## Out Of Scope

- Automatic route optimization.
- Historical route performance analytics beyond current health/latency summary.
- Editing provider credentials from the Routes page.
