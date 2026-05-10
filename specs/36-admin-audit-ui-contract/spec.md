# Feature Specification: Admin Audit UI Contract

**Feature Branch**: `[36-admin-audit-ui-contract]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Audit UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Filter Audit History for UI Investigations (Priority: P1)

Gateway administrators filter audit history using the same investigation concepts exposed by the admin UI: actor, action, date range, outcome, and pagination.

**Why this priority**: The Audit page already depends on a production audit endpoint, but its current query contract does not fully match the UI surface language. Aligning the contract prevents the UI from relying on mock-only filters or adapter-only behavior.

**Independent Test**: Seed audit records with multiple actors, actions, timestamps, and outcomes; request filtered audit history; verify returned rows, total count, and pagination metadata match the requested view.

**Acceptance Scenarios**:

1. **Given** audit records from multiple API keys or administrators, **When** an administrator filters by actor, **Then** every returned record belongs to that actor-like identity.
2. **Given** audit records across multiple request and admin actions, **When** an administrator filters by action, **Then** every returned record matches the action family or exact action requested.
3. **Given** a date range filter, **When** the endpoint returns audit history, **Then** every returned record falls within the inclusive requested range and results remain newest-first.
4. **Given** a paginated result set, **When** the UI requests the next page, **Then** the response includes stable pagination metadata and does not duplicate or skip records.

---

### User Story 2 - Preserve Safe Audit Detail Display (Priority: P2)

Security reviewers and operators can trust that audit rows exposed to the UI contain enough investigation context without leaking secrets or raw request bodies.

**Why this priority**: Audit data may include provider errors or administrative metadata. The UI requires safe summaries that remain useful during incident response.

**Independent Test**: Insert audit rows containing secret-like error text and metadata; request the rows through the admin audit contract; verify sensitive values are redacted or omitted.

**Acceptance Scenarios**:

1. **Given** an audit record has an error message containing an authorization header, token, credential, or raw API key, **When** it is returned for UI display, **Then** the sensitive value is redacted.
2. **Given** an audit record references a deleted key, model, provider, or actor, **When** the row is returned, **Then** the response still includes a stable identifier and a clear missing-reference indication.
3. **Given** audit records include request outcomes, **When** rows are returned, **Then** success, failure, unauthorized, and budget-denied outcomes can be distinguished by text, not color-only state.

### Edge Cases

- A filter references an actor, action, model, provider, or key that has no matching rows.
- The requested page is beyond the available result set.
- Date filters are malformed or the start date is after the end date.
- Records were created with older schema versions and have missing optional fields.
- Error text contains secret-like values in mixed case or nested structured data.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose audit history using actor, action, date range, outcome, limit, and page or offset filters that can directly support the admin Audit UI.
- **FR-002**: The system MUST return audit rows sorted newest-first with deterministic secondary ordering for records with equal timestamps.
- **FR-003**: The system MUST return pagination metadata that lets the UI preserve filters across page changes.
- **FR-004**: The system MUST preserve the existing audit query capabilities for API key, model, provider, status, date range, limit, and offset clients.
- **FR-005**: The system MUST redact or omit authorization headers, provider credentials, raw API keys, tokens, cookies, prompts, responses, and secret-like values before audit rows are displayed through the admin contract.
- **FR-006**: The system MUST represent missing referenced entities without failing the entire audit query.
- **FR-007**: The system MUST authenticate and authorize audit queries consistently with existing admin endpoint rules.
- **FR-008**: The system MUST preserve OpenAI-compatible gateway behavior and avoid changing request proxy semantics.
- **FR-009**: The system MUST provide success and failure tests for filtering, pagination, invalid input, redaction, and missing references.

### Key Entities *(include if feature involves data)*

- **Audit Record**: Immutable request or administrative activity with actor-like identity, action, request context, token and cost details, outcome, error summary, and timestamp.
- **Audit Filter**: User-selected constraints for actor, action, date range, outcome, pagination, and compatibility filters.
- **Audit Page Result**: A slice of audit records plus metadata describing total matches and current position.
- **Reference Summary**: Safe display information for related keys, models, providers, or actors, including deleted or unknown states.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can narrow seeded audit history by actor, action, and date range with 100% matching visible rows.
- **SC-002**: Paginated audit views preserve filters across at least three pages without duplicate or skipped rows.
- **SC-003**: Secret-like values are redacted in 100% of audit rows covered by security fixtures.
- **SC-004**: Existing clients using the current audit filters continue to receive compatible results.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing OpenAI-compatible client request and response behavior remains unchanged.
- **CC-002**: Existing admin audit consumers remain compatible or receive a documented migration path for any additive response fields.
- **CC-003**: Audit query failures return OpenAI-compatible admin errors without leaking database details.
- **CC-004**: Operators can verify UI audit behavior without direct database access.

## Assumptions

- The existing persisted audit log remains the source of truth for request and administrative history.
- Actor can map to an API key, administrator identity, service account, or future RBAC principal.
- Additive response fields are acceptable as long as current clients are not broken.
- Retention and deletion policy changes are out of scope for this spec.
