# Feature Specification: Admin Budgets Surface API

**Feature Branch**: `[41-admin-budgets-surface-api]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Budgets UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Monitor Budget Policies Across Scopes (Priority: P1)

Administrators view budget policies across team and key scopes with current usage, limit, duration, threshold, and hard-cap state.

**Why this priority**: Per-key budget APIs exist, but the Budgets UI needs an aggregate production contract that surfaces governance state across scopes.

**Independent Test**: Seed key and team budgets with healthy, near-cap, and over-cap states; list budgets; verify utilization, warning state, and scope context.

**Acceptance Scenarios**:

1. **Given** budgets exist for teams and keys, **When** an administrator lists budgets, **Then** each policy includes scope, display name, limit, usage, duration, alert threshold, and cap state.
2. **Given** a policy is at or above its alert threshold, **When** budgets are listed, **Then** the policy is marked near cap or over cap with an explicit state.
3. **Given** a key or team reference is missing, **When** budgets are listed, **Then** the policy remains visible with a missing-reference indication.
4. **Given** no budgets are configured, **When** budgets are listed, **Then** the response distinguishes no policies from budget service unavailability.

---

### User Story 2 - Update Budget Policy Without Losing Spend State (Priority: P1)

Administrators change a budget limit, duration, or alert threshold while preserving current spend usage and auditability.

**Why this priority**: The UI includes budget edit states, and production writes need contract-backed validation and state preservation.

**Independent Test**: Update a budget's limit, duration, and threshold after spend has accumulated; verify usage remains unchanged and audit history records the change.

**Acceptance Scenarios**:

1. **Given** a budget has existing usage, **When** the administrator changes its limit, **Then** the saved policy preserves current usage.
2. **Given** an invalid threshold, negative limit, or unsupported duration, **When** the change is submitted, **Then** it is rejected with a specific validation reason.
3. **Given** a valid budget update succeeds, **When** budgets are listed again, **Then** the updated policy and unchanged usage are visible.
4. **Given** a budget update succeeds or fails, **When** audit history is inspected, **Then** the actor, scope, prior policy, new policy, and outcome are traceable.

### Edge Cases

- A budget exists for a revoked API key.
- A team budget exists for an archived or deleted team.
- Usage exceeds the new limit immediately after a policy update.
- Alert threshold is missing, zero, above 100%, or malformed.
- Duration changes while a reset period is active.
- Two administrators update the same budget concurrently.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose aggregate budget policies across key and team scopes.
- **FR-002**: The system MUST include scope, scope id, safe display name, limit, current usage, duration, alert threshold, and cap state for each budget policy.
- **FR-003**: The system MUST classify budget health as healthy, near cap, over cap, or missing policy using explicit state values.
- **FR-004**: The system MUST support updating limit, duration, and alert threshold for supported budget scopes.
- **FR-005**: The system MUST preserve current usage when budget policy settings change.
- **FR-006**: The system MUST reject invalid limits, thresholds, durations, and unsupported scopes with specific validation errors.
- **FR-007**: The system MUST audit budget policy changes with actor, scope, prior policy, new policy, outcome, and timestamp.
- **FR-008**: The system MUST avoid returning raw API keys, credentials, or unrelated tenant data in budget responses.
- **FR-009**: The system MUST preserve existing budget enforcement behavior for client requests.
- **FR-010**: The system MUST provide tests for aggregate listing, warning states, updates, validation failures, usage preservation, missing references, and audit records.

### Key Entities *(include if feature involves data)*

- **Budget Policy**: Spend control scoped to a key or team with limit, usage, duration, threshold, and cap state.
- **Budget Scope**: The owner of a policy, such as a key or team, represented safely for admin display.
- **Budget Health State**: Derived classification of current usage relative to threshold and limit.
- **Budget Policy Change**: An administrator action that updates policy settings while preserving current spend state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can identify all seeded near-cap and over-cap budgets from one budget list query.
- **SC-002**: Budget policy updates preserve current usage in 100% of update fixtures.
- **SC-003**: Invalid budget updates return specific validation reasons in 100% of negative fixtures.
- **SC-004**: Successful and failed budget updates are traceable in audit history.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing per-key budget APIs and enforcement behavior remain compatible.
- **CC-002**: Requests over budget continue to be denied before provider calls according to existing spend governance rules.
- **CC-003**: Operators can manage budget policy state through admin APIs without direct database access.

## Assumptions

- Per-key budget enforcement already exists and remains the initial enforcement baseline.
- Team budgets may require team identity support from the Teams feature before full enforcement.
- Forecasting, real-time budget metering, and alert delivery are out of scope.
- Key-level one-off budget reset behavior remains covered by the existing budget management feature.
