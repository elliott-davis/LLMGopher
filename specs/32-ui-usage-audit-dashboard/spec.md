# Feature Specification: UI Usage and Audit Dashboard

**Feature Branch**: `[32-ui-usage-audit-dashboard]`  
**Created**: 2026-04-26  
**Status**: Draft  
**Input**: User description: "Close the UI/API gap for usage, spend, and audit visibility by surfacing existing admin analytics capabilities in the admin UI."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Review Usage and Spend Summary (Priority: P1)

Gateway administrators review request counts, token usage, cost, errors, and latency grouped by model, provider, or API key over a selected time window.

**Why this priority**: Usage and spend are core operational concerns, and the backend already exposes this data while the UI currently has no analytics page.

**Independent Test**: Open the usage dashboard, select each grouping option, adjust the time window, and verify the displayed summaries update accordingly.

**Acceptance Scenarios**:

1. **Given** usage data exists for the selected time window, **When** an administrator groups by model, **Then** the UI shows usage and spend totals for each model.
2. **Given** usage data exists for the selected time window, **When** an administrator groups by provider, **Then** the UI shows usage and spend totals for each provider.
3. **Given** usage data exists for the selected time window, **When** an administrator groups by API key, **Then** the UI shows usage and spend totals for each key.

---

### User Story 2 - Inspect Daily Trends (Priority: P2)

Gateway administrators view daily usage and spend trends to understand whether activity is increasing, decreasing, or spiking.

**Why this priority**: Daily trends help operators spot changes that are not obvious in aggregate totals.

**Independent Test**: Select a time window with multiple days of activity and verify that the UI presents one daily data point per day with request, token, and spend information.

**Acceptance Scenarios**:

1. **Given** a multi-day time window, **When** an administrator opens daily usage, **Then** the UI shows one entry per calendar day in the window.
2. **Given** a day with no usage, **When** the daily trend is shown, **Then** the UI handles the missing or zero day without breaking the trend view.

---

### User Story 3 - Search Audit Logs (Priority: P3)

Gateway administrators search request audit records by key, model, provider, status, and time range to investigate errors or unusual usage.

**Why this priority**: Audit records are essential for debugging and incident review, but the current UI gives operators no way to inspect them.

**Independent Test**: Apply audit filters and pagination in the UI, verify matching rows are shown, and inspect a row for request ID, model, provider, status, latency, token, cost, and error context.

**Acceptance Scenarios**:

1. **Given** audit records exist, **When** an administrator filters by error status, **Then** the UI shows only failed request records.
2. **Given** audit records span multiple pages, **When** an administrator changes pages, **Then** the UI shows the next set of matching records and preserves filters.
3. **Given** a request failed, **When** an administrator inspects its audit row, **Then** the UI shows useful debugging context without exposing secrets.

### Edge Cases

- No usage or audit records exist for the selected time window.
- The selected API key, model, or provider has been deleted but historical records remain.
- The gateway rejects an invalid time window or filter combination.
- Audit result totals exceed the page size limit.
- Usage and audit endpoints require administrator authentication that the UI session does not yet provide.
- Cost values are very small, zero, or rounded in a way that could mislead operators.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an admin UI view for usage and spend summaries.
- **FR-002**: The system MUST allow administrators to group usage by model, provider, and API key.
- **FR-003**: The system MUST allow administrators to filter usage summaries by time window and available dimensions.
- **FR-004**: The system MUST display request count, prompt tokens, completion tokens, total tokens, cost, error count, and average latency where available.
- **FR-005**: The system MUST provide a daily trend view for request count, token volume, and cost over the selected time window.
- **FR-006**: The system MUST provide an admin UI view for audit records.
- **FR-007**: The system MUST allow audit filtering by API key, model, provider, status, and time range.
- **FR-008**: The system MUST paginate audit results and communicate total result count when available.
- **FR-009**: The system MUST show audit row details needed for investigation, including request identifier, model, provider, token counts, cost, status, latency, streaming status, error context, and timestamp.
- **FR-010**: The system MUST handle empty, loading, unavailable, and invalid-filter states without losing the user's selected filters.
- **FR-011**: The system MUST avoid exposing API key secrets, provider credentials, or sensitive request payloads in usage or audit views.
- **FR-012**: The system MUST document any analytics or audit capability that remains API-only with a clear user-role rationale and follow-up trigger.

### Key Entities *(include if feature involves data)*

- **Usage Summary**: Aggregated request, token, cost, error, and latency information for a selected grouping and time window.
- **Daily Usage Point**: One day's request, token, and spend totals.
- **Audit Record**: A historical request record with routing, status, latency, token, cost, and error context.
- **Analytics Filter**: The selected time window, grouping, status, key, model, or provider used to narrow the displayed data.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can answer "which model, provider, or key drove the most spend in the last 30 days" in under 1 minute.
- **SC-002**: Administrators can identify whether request volume is trending up or down across a selected multi-day window in under 1 minute.
- **SC-003**: Administrators can find recent failed requests using audit filters in under 2 minutes.
- **SC-004**: Empty and unavailable states provide clear next steps 100% of the time.
- **SC-005**: Usage and audit views expose zero key secrets or provider credentials.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Usage summaries remain compatible with the backend usage capabilities defined by spec 08.
- **CC-002**: Audit search remains compatible with the backend audit capabilities defined by spec 06.
- **CC-003**: Operators can review usage, spend, and request history through the admin UI without issuing raw API calls.
- **CC-004**: The dashboard does not change request handling, billing, audit recording, or provider routing behavior.

## Assumptions

- Existing backend usage APIs from spec 08 and audit APIs from spec 06 are available and remain the source of truth.
- The UI is intended for trusted gateway administrators.
- Real-time streaming metrics, external analytics export, and long-term retention policy management remain out of scope for this UI parity feature.
- Historical records may refer to deleted keys, models, or providers, and should still remain inspectable.
