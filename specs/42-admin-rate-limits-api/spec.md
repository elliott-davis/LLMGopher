# Feature Specification: Admin Rate Limits API

**Feature Branch**: `[42-admin-rate-limits-api]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Rate Limits UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Central Rate-Limit Rules (Priority: P1)

Administrators view request and token throttling rules across model, key, and team scopes, including whether any rule is currently or recently tripped.

**Why this priority**: Existing rate-limit fields are scattered across keys and models; the UI needs a central production contract with scope context and tripped state.

**Independent Test**: Seed model, key, and team rate-limit rules with one tripped rule; list rate limits; verify scope, limits, and tripped indicator.

**Acceptance Scenarios**:

1. **Given** rate-limit rules exist across model, key, and team scopes, **When** an administrator lists rules, **Then** each rule includes scope, scope id, request limit, token limit when supported, and tripped state.
2. **Given** exactly one rule is tripped, **When** rules are listed, **Then** exactly one rule reports a tripped state.
3. **Given** token-per-minute enforcement is not available for a scope, **When** the rule is returned, **Then** token controls are marked unavailable instead of implied.
4. **Given** no rules are configured, **When** rules are listed, **Then** the response distinguishes empty configuration from rate-limit service unavailability.

---

### User Story 2 - Manage Rate-Limit Rules Safely (Priority: P2)

Administrators create, update, or delete rate-limit rules only where enforcement support exists and validation prevents ineffective or invalid rules.

**Why this priority**: The UI includes editable rate-limit states, but unsupported writes would create false confidence about enforcement.

**Independent Test**: Submit valid and invalid create, update, and delete operations; verify persistence, validation, enforcement availability, and audit records.

**Acceptance Scenarios**:

1. **Given** a supported scope, **When** an administrator creates a rule with a positive request limit, **Then** the rule appears in subsequent reads.
2. **Given** a negative request or token limit, **When** a rule is submitted, **Then** it is rejected with a specific validation reason.
3. **Given** no meaningful limit is configured, **When** a rule is submitted, **Then** it is rejected as ineffective.
4. **Given** a rule is deleted successfully, **When** rules are listed again, **Then** the deleted rule no longer appears and the deletion is auditable.

### Edge Cases

- A rule references a deleted key, model, or team.
- Multiple rules apply to the same request.
- A rule is tripped while an administrator edits it.
- Token limits are configured for a scope where token enforcement is unavailable.
- A delete request targets an already-deleted rule.
- Team-level limits exist before team membership is fully integrated.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose central rate-limit rules across model, key, and team scopes.
- **FR-002**: The system MUST include scope, scope id, request-per-second limit, token-per-minute limit when supported, enforcement availability, and tripped state for each rule.
- **FR-003**: The system MUST distinguish tripped, not tripped, unknown, and unavailable states using explicit response values.
- **FR-004**: The system MUST support creating, updating, and deleting rules only for scopes with confirmed enforcement support.
- **FR-005**: The system MUST reject negative limits, missing meaningful limits, unsupported scopes, and unsupported token enforcement with specific validation errors.
- **FR-006**: The system MUST make successful rule changes visible in subsequent reads.
- **FR-007**: The system MUST audit rate-limit rule changes with actor, rule identity, prior rule, new rule, outcome, and timestamp.
- **FR-008**: The system MUST preserve existing model-level and key-level rate-limit behavior unless explicitly changed by a supported admin operation.
- **FR-009**: The system MUST provide tests for listing, tripped state, CRUD operations, validation failures, unsupported token limits, missing references, and audit records.

### Key Entities *(include if feature involves data)*

- **Rate-Limit Rule**: A throttling policy scoped to a model, key, or team.
- **Rate-Limit Scope**: The target entity a rule applies to, represented safely for admin display.
- **Tripped State**: Current or recent indication that a rule limited traffic.
- **Enforcement Capability**: Whether request or token throttling is actually supported for the rule's scope.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can identify every seeded tripped rule from one rate-limit list query.
- **SC-002**: Invalid or unsupported rule writes return specific validation reasons in 100% of negative fixtures.
- **SC-003**: Successful rule create, update, and delete operations are reflected in subsequent reads.
- **SC-004**: Token-per-minute controls are marked unavailable in 100% of scopes where enforcement is not supported.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing request throttling behavior remains compatible for models and keys.
- **CC-002**: No UI-visible rule implies enforcement unless enforcement support exists.
- **CC-003**: Rate-limit decisions remain observable through audit, logs, or rule state.

## Assumptions

- Existing model and key request-per-second limits are the starting point for production support.
- Token-per-minute enforcement may require separate runtime support before writes are enabled.
- Historical rate-limit analytics and automatic remediation are out of scope.
- Team-level enforcement depends on team identity support from the Teams feature.
