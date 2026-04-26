# Feature Specification: UI API Key Budget Controls

**Feature Branch**: `[31-ui-key-budgets]`  
**Created**: 2026-04-26  
**Status**: Draft  
**Input**: User description: "Close the UI/API gap for API key budgets by surfacing existing budget management capabilities in the admin UI."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Key Budget State (Priority: P1)

Gateway administrators inspect each API key's budget limit, spend, remaining balance, alert threshold, duration, and reset timing from the admin UI.

**Why this priority**: Spend controls already exist in the backend, but operators cannot see budget state without raw API calls or database access.

**Independent Test**: Open an API key's budget view and verify the displayed limit, spend, remaining balance, and reset information match the gateway state.

**Acceptance Scenarios**:

1. **Given** a key with a configured budget, **When** an administrator opens its budget details, **Then** the UI shows the budget limit, amount spent, amount remaining, and reset policy.
2. **Given** a key without a configured budget, **When** an administrator opens its budget details, **Then** the UI clearly states that no budget is set and offers a setup action.

---

### User Story 2 - Set or Update Budgets (Priority: P2)

Gateway administrators set or update per-key budgets, alert thresholds, duration, and reset timing from the UI.

**Why this priority**: Operators need to manage spend guardrails in the same place where credentials are managed.

**Independent Test**: Configure a budget for a key through the UI, save it, refresh the budget view, and verify the new budget settings are displayed.

**Acceptance Scenarios**:

1. **Given** a key without a budget, **When** an administrator sets a valid budget, **Then** the UI shows the saved budget state.
2. **Given** a key with an existing budget, **When** an administrator updates the limit or threshold, **Then** the UI shows the updated values without resetting spend unless explicitly requested.
3. **Given** a duration-based budget is selected, **When** an administrator saves the budget, **Then** the UI requires and displays the associated reset time.

---

### User Story 3 - Reset or Remove Budgets (Priority: P3)

Gateway administrators reset budget spend counters or remove budget limits when business policy changes.

**Why this priority**: Reset and removal actions are less frequent but important for operational recovery and policy changes.

**Independent Test**: Reset spend for a budgeted key and remove a budget from a test key, verifying the UI reflects each resulting state.

**Acceptance Scenarios**:

1. **Given** a key with non-zero spend, **When** an administrator resets budget spend, **Then** the UI shows spent amount as zero and recalculates remaining budget.
2. **Given** a key with a configured budget, **When** an administrator removes the budget after confirmation, **Then** the UI returns to the no-budget state.

### Edge Cases

- A budget is removed by another administrator while the current user is editing it.
- A reset action succeeds but the refreshed budget state is delayed.
- A budget value is zero, negative, or otherwise invalid.
- A threshold is outside the accepted range.
- A duration is selected without a reset time.
- The budget API requires administrator authentication that the UI session does not yet provide.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow administrators to view budget state for an API key from the admin UI.
- **FR-002**: The system MUST clearly distinguish keys with configured budgets from keys without budgets.
- **FR-003**: The system MUST allow administrators to set and update budget limit, alert threshold, duration, and reset timing for a key.
- **FR-004**: The system MUST preserve current spend when updating budget configuration unless the administrator explicitly resets spend.
- **FR-005**: The system MUST allow administrators to reset spent amount for a key budget with explicit confirmation.
- **FR-006**: The system MUST allow administrators to remove a key budget with explicit confirmation.
- **FR-007**: The system MUST show budget progress in a way that communicates spent and remaining amounts.
- **FR-008**: The system MUST prevent invalid budget submissions and explain the validation issue to the administrator.
- **FR-009**: The system MUST refresh budget state after set, update, reset, and remove actions.
- **FR-010**: The system MUST document any budget lifecycle capability that remains API-only with a clear user-role rationale and follow-up trigger.

### Key Entities *(include if feature involves data)*

- **API Key Budget**: The spend control attached to one API key, including limit, spent amount, remaining amount, alert threshold, duration, and reset time.
- **Budget Policy**: The operator-selected configuration that determines how much spend is allowed and when it resets.
- **Budget Lifecycle Action**: A set, update, reset, or remove operation performed by an administrator.
- **Budget Status Indicator**: UI feedback that communicates whether a key is unbudgeted, within budget, near threshold, or exhausted.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can determine whether a key has a budget and how much remains in under 30 seconds.
- **SC-002**: Administrators can configure a valid key budget from the UI in under 2 minutes.
- **SC-003**: At least 95% of successful budget changes show updated budget state after the expected refresh window.
- **SC-004**: Reset and remove actions require explicit confirmation 100% of the time.
- **SC-005**: Invalid budget submissions are rejected before or during save with clear, actionable feedback.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing budget behavior remains compatible with the backend budget management capabilities defined by spec 07.
- **CC-002**: Budget lifecycle controls do not change request enforcement semantics for API clients beyond the administrator's explicit budget changes.
- **CC-003**: Operators can configure, view, and validate key budgets through the admin UI without issuing raw API calls.

## Assumptions

- Existing backend budget APIs from spec 07 are available and remain the source of truth.
- Budget lifecycle scheduling and alert delivery beyond the existing budget fields remain governed by the budget lifecycle roadmap.
- The UI is intended for trusted gateway administrators.
- Per-team and per-model budgets remain out of scope for this UI parity feature.
