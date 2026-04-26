# Feature Specification: UI API Key Lifecycle Controls

**Feature Branch**: `[30-ui-key-lifecycle]`  
**Created**: 2026-04-26  
**Status**: Draft  
**Input**: User description: "Close the UI/API gap for API key lifecycle controls by surfacing existing key management capabilities in the admin UI."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Update Existing Keys (Priority: P1)

Gateway administrators update API key names, rate limits, expiration, model access, metadata, and active status from the admin UI without issuing raw API calls or editing the database.

**Why this priority**: The backend already supports mutable key configuration, but the UI only lists and creates keys. Operators need lifecycle controls to safely manage production credentials.

**Independent Test**: Create or select an API key in the UI, update each editable field, refresh the key inventory, and verify the updated values are visible and preserved.

**Acceptance Scenarios**:

1. **Given** an existing active API key, **When** an administrator changes its name and rate limit, **Then** the key inventory shows the new values after save.
2. **Given** an existing API key, **When** an administrator sets an expiration time, **Then** the key inventory shows the expiration state and the saved value remains available after refresh.
3. **Given** an existing API key, **When** an administrator restricts it to selected models, **Then** the UI clearly shows the allowed model list for that key.

---

### User Story 2 - Disable or Re-enable Keys (Priority: P2)

Gateway administrators deactivate compromised or unused keys and reactivate keys when needed.

**Why this priority**: Deactivation is safer than immediate deletion when operators need a reversible action during incident response or rollout.

**Independent Test**: Toggle a key between active and inactive in the UI and verify the status is updated without creating a new key.

**Acceptance Scenarios**:

1. **Given** an active key, **When** an administrator deactivates it, **Then** the UI shows the key as inactive and presents a reactivation option.
2. **Given** an inactive key, **When** an administrator reactivates it, **Then** the UI shows the key as active again.

---

### User Story 3 - Delete Keys (Priority: P3)

Gateway administrators permanently delete keys that are no longer needed, with clear confirmation and consequences.

**Why this priority**: Deletion is a destructive lifecycle action and should be available only with intentional confirmation.

**Independent Test**: Delete a non-production test key from the UI, confirm the warning, and verify the key is removed from the inventory after cache synchronization.

**Acceptance Scenarios**:

1. **Given** an existing key, **When** an administrator chooses delete, **Then** the UI displays a confirmation explaining that the action is permanent.
2. **Given** deletion is confirmed, **When** the gateway accepts the delete operation, **Then** the UI removes the key from the inventory after synchronization.

### Edge Cases

- The gateway is temporarily unavailable while saving key changes.
- A key disappears because another administrator deleted it before the current user saves changes.
- The selected allowed model list becomes stale because models changed while the key editor was open.
- Invalid metadata input is rejected with a clear message and without losing the rest of the form state.
- A deletion request is accepted but cache synchronization takes longer than expected.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow administrators to edit existing API key name, rate limit, expiration, metadata, allowed models, and active status from the admin UI.
- **FR-002**: The system MUST show current lifecycle fields for each key in a way that lets administrators understand whether a key is active, expired, restricted, or unrestricted.
- **FR-003**: The system MUST provide a reversible active/inactive control for API keys.
- **FR-004**: The system MUST provide a permanent delete action for API keys with explicit confirmation before the action is submitted.
- **FR-005**: The system MUST preserve the existing one-time secret display behavior for newly created keys.
- **FR-006**: The system MUST allow key creation to include the same lifecycle fields that can be edited later where those fields are known at creation time.
- **FR-007**: The system MUST present model allowlists using human-readable model information when available, while preserving the exact model identifiers needed for enforcement.
- **FR-008**: The system MUST surface validation and gateway errors in the UI without exposing secret key material.
- **FR-009**: The system MUST refresh key state after create, update, deactivate/reactivate, and delete actions so the UI reflects gateway cache synchronization.
- **FR-010**: The system MUST document any key lifecycle capability that remains API-only with a clear user-role rationale and follow-up trigger.

### Key Entities *(include if feature involves data)*

- **API Key**: A managed gateway credential with name, status, rate limit, expiration, metadata, allowed models, and creation/update timestamps.
- **Model Access Rule**: The list of model identifiers a key may use; an empty list means unrestricted access.
- **Key Lifecycle Action**: A create, update, deactivate, reactivate, or delete operation performed by an administrator.
- **Operator Feedback**: UI messages, warnings, and refresh states that communicate whether an action succeeded, failed, or is waiting for gateway synchronization.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can update all key lifecycle fields from the UI in under 2 minutes for an existing key.
- **SC-002**: At least 95% of successful key lifecycle actions show the updated state in the UI after the expected cache synchronization window.
- **SC-003**: Administrators can deactivate a key without deleting it and reactivate it later without creating a replacement key.
- **SC-004**: Destructive delete actions require explicit confirmation 100% of the time.
- **SC-005**: User-facing errors for failed key operations identify the failed action and next step without exposing key secrets.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing API key behavior remains compatible with the backend key management capabilities defined by spec 05.
- **CC-002**: Existing gateway clients continue to authenticate the same way unless an administrator explicitly changes or deletes their key.
- **CC-003**: Operators can configure, view, and validate key lifecycle state through the admin UI without issuing raw API calls.

## Assumptions

- Existing backend key lifecycle APIs from spec 05 are available and remain the source of truth.
- Existing model inventory is sufficient to populate allowed model choices.
- The UI is intended for trusted gateway administrators.
- Key rotation that preserves the same key ID remains out of scope for this UI parity feature.
