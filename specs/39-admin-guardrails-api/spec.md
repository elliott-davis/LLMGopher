# Feature Specification: Admin Guardrails API

**Feature Branch**: `[39-admin-guardrails-api]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Guardrails UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Guardrail Policies (Priority: P1)

Administrators view configured guardrail policies, including identity, category, enabled state, description, provider label, and last update information.

**Why this priority**: The UI must reflect real safety controls rather than mock-only fixtures, and operators need to confirm policy posture during incidents.

**Independent Test**: Seed guardrail policies across prompt, response, PII, jailbreak, and secret-scanning categories; query the list; verify safe fields and enabled states.

**Acceptance Scenarios**:

1. **Given** guardrail policies exist, **When** an administrator lists guardrails, **Then** each policy includes display name, enabled state, category, description, provider label, and last updated time when available.
2. **Given** no guardrails are configured, **When** guardrails are listed, **Then** the response distinguishes an empty policy set from service unavailability.
3. **Given** guardrail metadata includes sensitive detector payloads, **When** guardrails are listed, **Then** raw prompts, matches, credentials, and provider secrets are not returned.

---

### User Story 2 - Toggle Guardrails Safely (Priority: P1)

Administrators enable or disable an existing guardrail and can verify that the persisted state affects subsequent gateway policy checks.

**Why this priority**: The Guardrails UI acceptance criteria include toggle persistence, but production toggles require a real backend contract and audit trail.

**Independent Test**: Toggle a seeded guardrail on and off; reload state; send a request affected by that guardrail; verify state persistence, auditability, and predictable failure behavior.

**Acceptance Scenarios**:

1. **Given** a disabled guardrail, **When** an authorized administrator enables it, **Then** subsequent reads show it enabled.
2. **Given** an enabled guardrail, **When** an authorized administrator disables it, **Then** subsequent reads show it disabled.
3. **Given** a toggle request fails validation or storage persistence, **When** the response is returned, **Then** the final state remains unambiguous.
4. **Given** a guardrail state changes, **When** audit history is inspected, **Then** the actor, guardrail id, prior state, new state, and outcome are traceable.

### Edge Cases

- A guardrail id does not exist.
- A guardrail provider integration is unavailable.
- A toggle request races with another administrator's toggle.
- A guardrail contains provider credentials or detector rules that must not be displayed.
- A disabled guardrail was previously referenced by audit records.
- An administrator lacks permission to mutate guardrail state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose guardrail policies with safe display fields and enabled state.
- **FR-002**: The system MUST support enabling and disabling existing guardrails through an authenticated and authorized admin operation.
- **FR-003**: The system MUST persist guardrail enabled state so subsequent reads and policy checks observe the saved state.
- **FR-004**: The system MUST make failed toggle operations leave the final persisted state unambiguous.
- **FR-005**: The system MUST audit guardrail state changes with actor, guardrail identity, prior state, new state, outcome, and timestamp.
- **FR-006**: The system MUST omit raw detector prompts, matched prompt/response content, provider credentials, and sensitive policy internals from admin responses.
- **FR-007**: The system MUST distinguish empty guardrail configuration from unavailable guardrail storage or integrations.
- **FR-008**: The system MUST preserve OpenAI-compatible gateway behavior except for administrator-approved guardrail state changes.
- **FR-009**: The system MUST provide tests for listing, toggle success, toggle failure, reload persistence, policy-check effect, audit records, and redaction.

### Key Entities *(include if feature involves data)*

- **Guardrail Policy**: A safety control with identity, category, enabled state, provider label, description, and update metadata.
- **Guardrail State Change**: An administrator action that changes enabled state and produces audit evidence.
- **Guardrail Integration Status**: Safe operator-facing availability information for the mechanism backing the policy.
- **Sensitive Detector Payload**: Internal detector or match data that must never be rendered in admin UI responses.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can see current enabled state for every configured guardrail from one list query.
- **SC-002**: A successful guardrail toggle is visible after reload in 100% of toggle persistence tests.
- **SC-003**: Failed toggle operations preserve or restore a clearly reported final state in 100% of failure fixtures.
- **SC-004**: Sensitive detector and credential values are absent from 100% of guardrail list and mutation responses.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing client request schemas remain unchanged.
- **CC-002**: Guardrail state changes are auditable and visible to operators.
- **CC-003**: Guardrail enforcement remains in the request path only where configured by administrators.
- **CC-004**: Admin guardrail failures return compatible errors without exposing provider or detector secrets.

## Assumptions

- Guardrail creation and detailed detector configuration are out of scope for the first production admin contract.
- Existing guardrail middleware remains responsible for request enforcement.
- RBAC-specific permission names may be introduced by a separate authorization spec.
- Sensitive matched prompt and response content remains out of scope for display.
