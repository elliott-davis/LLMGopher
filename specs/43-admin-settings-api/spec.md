# Feature Specification: Admin Settings API

**Feature Branch**: `[43-admin-settings-api]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Settings UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Review Organization Settings (Priority: P1)

Administrators review bounded organization settings for Gateway Profile, Security, Notifications, and Display without exposing secrets or raw configuration.

**Why this priority**: The Settings UI should not become a catch-all or display unsafe runtime configuration. A production backend contract must define safe settings groups before editable controls ship.

**Independent Test**: Seed safe setting cards with read-only, editable, and unavailable fields; query settings; verify required cards and redaction.

**Acceptance Scenarios**:

1. **Given** settings are available, **When** an administrator opens settings, **Then** Gateway Profile, Security, Notifications, and Display groups are returned.
2. **Given** a setting has no backend support, **When** settings are returned, **Then** it is marked unavailable or read-only with explanatory copy.
3. **Given** a setting value contains a secret-like value, **When** settings are returned, **Then** the value is redacted or omitted.
4. **Given** a settings source is unavailable, **When** settings are queried, **Then** unavailable groups are distinguished from intentionally empty fields.

---

### User Story 2 - Save Supported Settings Safely (Priority: P2)

Administrators update only settings with confirmed backend support while validation, failure behavior, and audit records preserve confidence.

**Why this priority**: Settings changes can affect authentication posture, notifications, or user experience. Unsupported settings must not appear writable.

**Independent Test**: Submit supported and unsupported setting changes; verify validation, persistence, unavailable behavior, and audit history.

**Acceptance Scenarios**:

1. **Given** an editable setting with backend support, **When** an administrator saves a valid value, **Then** subsequent reads show the persisted value.
2. **Given** a read-only or unavailable setting, **When** an administrator attempts to save it, **Then** the system rejects the change with a clear reason.
3. **Given** validation fails, **When** the save response is returned, **Then** the invalid field is identified and no unrelated settings change.
4. **Given** a setting change succeeds or fails, **When** audit history is inspected, **Then** the actor, card, field, prior safe value, new safe value, and outcome are traceable.

### Edge Cases

- A setting value is secret-like and cannot safely be echoed after save.
- A setting group is available but one field is unavailable.
- Notification settings are configured without a delivery backend.
- Security posture is read-only because authentication is configured outside the gateway.
- A local UI preference exists that should not be stored as organization-wide state.
- Two administrators edit the same setting concurrently.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose bounded settings groups for Gateway Profile, Security, Notifications, and Display.
- **FR-002**: The system MUST represent each setting group and field with availability state: read-only, editable, or unavailable.
- **FR-003**: The system MUST include explanatory copy for unavailable or read-only backend-dependent controls.
- **FR-004**: The system MUST redact or omit provider credentials, admin API keys, raw configuration values, tokens, cookies, and secret-like values.
- **FR-005**: The system MUST support saving only settings with confirmed backend support.
- **FR-006**: The system MUST reject writes to read-only, unavailable, unknown, or unsupported settings with clear validation reasons.
- **FR-007**: The system MUST preserve unrelated setting values when one field fails validation or persistence.
- **FR-008**: The system MUST audit supported setting changes with actor, group, field, prior safe value, new safe value, outcome, and timestamp.
- **FR-009**: The system MUST preserve existing configuration precedence and startup behavior unless a setting is explicitly designed to change runtime state.
- **FR-010**: The system MUST provide tests for all four groups, read-only states, unavailable states, successful saves, validation failures, secret redaction, and audit records.

### Key Entities *(include if feature involves data)*

- **Settings Group**: A bounded collection of settings for Gateway Profile, Security, Notifications, or Display.
- **Setting Field**: One setting value with label, availability, validation state, and safe display value.
- **Setting Change**: An administrator action that updates a supported field and creates audit evidence.
- **Secret-Like Value**: Any value that must be redacted or omitted from settings display and audit summaries.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can retrieve all four required settings groups from one settings query.
- **SC-002**: Unsupported settings are marked read-only or unavailable in 100% of backend-dependent fixtures.
- **SC-003**: Secret-like settings values are redacted or omitted in 100% of security fixtures.
- **SC-004**: Supported setting saves are visible in subsequent reads and auditable.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing gateway configuration behavior remains unchanged unless a setting explicitly supports runtime updates.
- **CC-002**: Settings responses never expose provider credentials, admin API keys, or raw config file values.
- **CC-003**: Operators can distinguish organization-wide settings from local-only UI preferences.

## Assumptions

- Provider credential management, admin API key rotation, and full RBAC role editing remain out of scope.
- Local-only UI preferences may remain client-side and do not require organization-wide persistence.
- Notification delivery backend implementation is out of scope unless a separate feature introduces it.
- Runtime configuration precedence remains governed by existing configuration rules.
