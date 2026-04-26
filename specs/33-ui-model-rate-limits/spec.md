# Feature Specification: UI Model Rate Limit Controls

**Feature Branch**: `[33-ui-model-rate-limits]`  
**Created**: 2026-04-26  
**Status**: Draft  
**Input**: User description: "Close the UI/API gap for per-model rate limits by surfacing model rate limit configuration in the admin UI."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Model Rate Limits (Priority: P1)

Gateway administrators set a per-model request rate limit when creating or editing a model in the admin UI.

**Why this priority**: The backend supports model-specific limits, but current model forms omit this field, leaving operators unable to configure it from the UI.

**Independent Test**: Create a model with a rate limit, edit that limit, refresh the model list, and verify the configured value is visible.

**Acceptance Scenarios**:

1. **Given** an administrator is creating a model, **When** they enter a non-negative model rate limit and save, **Then** the model inventory shows the configured limit.
2. **Given** an existing model, **When** an administrator changes its model rate limit, **Then** the model inventory shows the updated limit after save.
3. **Given** an administrator sets a model rate limit to zero, **When** the model is saved, **Then** the UI communicates that no model-level limit is configured.

---

### User Story 2 - Understand Model Throttle Policy (Priority: P2)

Gateway administrators can quickly see which models have additional throttling and which are unrestricted at the model level.

**Why this priority**: Operators need at-a-glance visibility into expensive or sensitive model controls.

**Independent Test**: Open the model inventory with a mix of limited and unlimited models and verify each row communicates its model-level policy clearly.

**Acceptance Scenarios**:

1. **Given** a model has a positive rate limit, **When** the model inventory is viewed, **Then** the UI shows the limit value.
2. **Given** a model has no model-level limit, **When** the model inventory is viewed, **Then** the UI shows an unrestricted or no-limit state.

---

### User Story 3 - Handle Invalid Rate Limit Input (Priority: P3)

Gateway administrators receive clear feedback when a model rate limit value is invalid.

**Why this priority**: Invalid throttling settings can lead to confusing save failures or accidental policy gaps.

**Independent Test**: Attempt to save invalid model rate limit values and verify the UI prevents or explains the failure.

**Acceptance Scenarios**:

1. **Given** an administrator enters a negative rate limit, **When** they try to save the model, **Then** the UI rejects the value with a clear message.
2. **Given** the gateway rejects a model rate limit update, **When** the save fails, **Then** the UI keeps the form state and displays the failure reason.

### Edge Cases

- Existing models do not yet have a rate limit value populated.
- Another administrator changes a model's rate limit while the edit form is open.
- The gateway is unavailable while creating or updating a model.
- A rate limit update is saved but cache synchronization takes longer than expected.
- Operators confuse key-level and model-level rate limits.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow administrators to set a non-negative model rate limit when creating a model.
- **FR-002**: The system MUST allow administrators to update a model's rate limit when editing a model.
- **FR-003**: The system MUST show model rate limit state in the model inventory.
- **FR-004**: The system MUST clearly distinguish no model-level limit from a positive model-level limit.
- **FR-005**: The system MUST prevent or reject negative model rate limit values with clear feedback.
- **FR-006**: The system MUST preserve existing model create, edit, delete, provider assignment, alias, name, and context window behavior.
- **FR-007**: The system MUST explain that model-level limits are separate from API key rate limits.
- **FR-008**: The system MUST refresh model state after create and update actions so the UI reflects gateway cache synchronization.
- **FR-009**: The system MUST document any model throttling capability that remains API-only with a clear user-role rationale and follow-up trigger.

### Key Entities *(include if feature involves data)*

- **Model Configuration**: A gateway model alias and its associated provider, display name, context window, and model-level rate limit.
- **Model Rate Limit**: The maximum allowed request rate for one model; zero means no model-level limit.
- **Rate Limit Policy Explanation**: User-facing copy that clarifies how model limits relate to key-level limits.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can configure a model rate limit during model creation in under 1 minute.
- **SC-002**: Administrators can identify which configured models have model-level limits in under 30 seconds.
- **SC-003**: Invalid negative rate limit values are rejected 100% of the time.
- **SC-004**: Existing model management workflows continue to succeed with no extra required fields beyond the model rate limit default.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Model rate limit controls remain compatible with the backend per-model rate limit behavior defined by spec 09.
- **CC-002**: Existing clients continue to call models the same way; only administrator-configured throttling behavior changes.
- **CC-003**: Operators can configure and inspect model rate limits through the admin UI without issuing raw API calls.

## Assumptions

- Existing backend per-model rate limit behavior from spec 09 is available and remains the source of truth.
- A rate limit value of zero means no model-level rate limit.
- The UI is intended for trusted gateway administrators.
- Token-per-minute limits and per-key-per-model compound limits remain out of scope for this UI parity feature.
