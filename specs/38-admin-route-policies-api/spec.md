# Feature Specification: Admin Route Policies API

**Feature Branch**: `[38-admin-route-policies-api]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Routes UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Inspect Routing Policies (Priority: P1)

Gateway administrators inspect how model traffic is routed across providers, including single-provider, fallback, weighted, and latency-aware policies.

**Why this priority**: The Routes UI needs production route policy data rather than inferring behavior from models and providers alone.

**Independent Test**: Configure representative routing policies; query them through the admin surface; verify each route includes strategy, provider targets, enabled state, and health context.

**Acceptance Scenarios**:

1. **Given** configured model routing policies, **When** an administrator opens the route policy list, **Then** each route shows model alias, strategy, enabled state, provider targets, and current policy summary.
2. **Given** a fallback route, **When** the route is returned, **Then** it identifies a primary provider and at least one fallback provider in deterministic order.
3. **Given** a weighted route, **When** the route is returned, **Then** every target has a non-negative weight and the route identifies how traffic is distributed.
4. **Given** a latency-aware route, **When** the route is returned, **Then** operators can see the health or latency context used for decision-making.

---

### User Story 2 - Change Routing Policies Safely (Priority: P2)

Administrators update routing policies only through validated changes that avoid invalid provider mixes, missing primary providers, and zero-total weighted routes.

**Why this priority**: Route changes can affect all client traffic, so writes must be explicit, validated, auditable, and reversible.

**Independent Test**: Submit valid and invalid route changes; verify invalid changes are rejected with specific reasons and valid changes are reflected in subsequent route reads.

**Acceptance Scenarios**:

1. **Given** an administrator submits a valid fallback route update, **When** the change is saved, **Then** subsequent reads show the new primary and fallback order.
2. **Given** a weighted route update with negative weights or zero total weight, **When** the change is submitted, **Then** it is rejected with a specific validation reason.
3. **Given** a route references an unknown or disabled provider, **When** the change is submitted, **Then** the system prevents traffic from being assigned to the invalid target.
4. **Given** a route update succeeds, **When** audit history is inspected, **Then** the route change is traceable to the actor and changed policy.

### Edge Cases

- A route references a provider that has been deleted or disabled.
- A model alias has no route policy and falls back to existing registry behavior.
- A weighted policy sums to zero after a target is removed.
- A fallback chain includes duplicate providers.
- A route change is submitted while provider health changes.
- An administrator attempts to edit a route without sufficient permission.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose route policies with model alias, strategy, enabled state, provider targets, fallback order, weights, and health context.
- **FR-002**: The system MUST support single-provider, fallback, weighted, and latency-aware routing policy representations.
- **FR-003**: The system MUST validate route policy changes before they affect client traffic.
- **FR-004**: The system MUST reject missing primary providers, duplicate fallback providers, negative weights, zero-total weighted routes, unknown providers, and disabled provider targets.
- **FR-005**: The system MUST make successful route changes visible in subsequent admin reads.
- **FR-006**: The system MUST audit route policy changes with actor, route identity, before/after summary, outcome, and timestamp.
- **FR-007**: The system MUST preserve existing model and provider behavior for routes that are not explicitly changed.
- **FR-008**: The system MUST authenticate and authorize route policy reads and writes consistently with admin endpoint rules.
- **FR-009**: The system MUST preserve OpenAI-compatible client behavior except for administrator-approved routing policy changes.
- **FR-010**: The system MUST provide tests for route listing, each strategy type, validation failures, successful updates, audit records, and missing references.

### Key Entities *(include if feature involves data)*

- **Route Policy**: A model-to-provider routing rule with strategy, enabled state, provider targets, and safety metadata.
- **Route Target**: A provider participating in a route with order, weight, health state, and latency context.
- **Route Change**: An administrator-requested policy update with validation result and audit trail.
- **Provider Health Summary**: Operator-facing status used to explain route behavior without replacing runtime health checks.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can identify the current strategy and provider targets for every configured route from one route policy query.
- **SC-002**: Invalid route policy updates are rejected with specific reasons in 100% of validation fixtures.
- **SC-003**: Successful route policy updates are reflected in the next admin read and recorded in audit history.
- **SC-004**: Existing client requests continue to route according to prior behavior until an administrator saves a validated policy change.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing OpenAI-compatible client request and response schemas remain unchanged.
- **CC-002**: Routing changes are observable through admin reads and audit records.
- **CC-003**: Provider credentials are never exposed through route policy responses.
- **CC-004**: Operators can inspect route readiness without editing provider credentials or model definitions.

## Assumptions

- Existing provider and model records remain the source of truth for available route targets.
- Provider credential management remains outside the Routes surface.
- Automatic route optimization is out of scope.
- Historical route analytics are out of scope beyond current health and latency summary.
