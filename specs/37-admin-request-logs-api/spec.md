# Feature Specification: Admin Request Logs API

**Feature Branch**: `[37-admin-request-logs-api]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Logs UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Review Recent Gateway Requests (Priority: P1)

Gateway operators view recent request log rows with status, latency, key, model, path, and routing outcome so they can investigate failures without direct database access.

**Why this priority**: The Logs UI cannot be production-backed until the gateway exposes request-focused records distinct from aggregate audit history.

**Independent Test**: Generate successful, client-error, server-error, and fallback-handled requests; query request logs; verify rows contain safe investigation fields and newest-first ordering.

**Acceptance Scenarios**:

1. **Given** recent gateway requests with mixed outcomes, **When** an operator queries request logs, **Then** rows include request id, timestamp, method, path, status, latency, safe key label, model, and routing outcome.
2. **Given** a status-group filter, **When** an operator filters by success, client error, server error, or fallback, **Then** every returned row matches the selected group.
3. **Given** more rows than the first page, **When** the operator pages through results, **Then** pagination remains stable and newest-first.
4. **Given** no rows match a filter, **When** logs are queried, **Then** the response distinguishes an empty result from service unavailability.

---

### User Story 2 - Inspect a Request Safely (Priority: P1)

Operators open a specific request log and inspect provider trace, headers, prompt preview, and response preview without exposing raw secrets or full payloads.

**Why this priority**: The Logs UI request inspector is a core operational workflow and needs backend-provided safe detail rather than mock-only data.

**Independent Test**: Create a request with fallback trace, secret-bearing headers, and prompt/response content; fetch its detail; verify trace accuracy and redaction.

**Acceptance Scenarios**:

1. **Given** a fallback-handled request, **When** the operator opens its detail, **Then** the trace identifies the failed primary provider and the successful fallback provider.
2. **Given** request headers contain authorization, cookie, token, credential, or secret-like values, **When** detail is returned, **Then** those values are redacted.
3. **Given** prompt or response content exists, **When** detail is returned, **Then** only redacted, truncated previews are available.
4. **Given** a request detail cannot be found, **When** the operator opens it, **Then** the response uses a predictable not-found error without leaking storage details.

### Edge Cases

- A request has no provider call because authentication, budget, rate limit, or guardrail checks denied it.
- A provider attempt times out before any response body is available.
- A fallback chain has more than two provider attempts.
- A request references a deleted key, model, or provider.
- The captured request body is larger than the configured preview length.
- Secret-like values appear in nested metadata, headers, provider errors, prompts, or responses.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose paginated request log rows for recent gateway traffic.
- **FR-002**: The system MUST support filters for status group, fallback outcome, limit, and pagination.
- **FR-003**: The system MUST expose safe request detail for a single log row, including provider trace, redacted headers, prompt preview, and response preview.
- **FR-004**: The system MUST identify provider-chain outcomes including success, failed, skipped, and unavailable stages.
- **FR-005**: The system MUST redact or omit authorization headers, cookies, provider credentials, raw API keys, tokens, and secret-like values from all log responses.
- **FR-006**: The system MUST return prompt and response content only as bounded previews suitable for display.
- **FR-007**: The system MUST include enough missing-reference context for deleted keys, models, or providers without failing the request log query.
- **FR-008**: The system MUST authenticate and authorize request log access consistently with existing admin endpoint rules.
- **FR-009**: The system MUST preserve OpenAI-compatible gateway behavior and avoid changing provider response semantics.
- **FR-010**: The system MUST provide tests for list filters, detail lookup, fallback traces, empty results, unavailable storage, and redaction.

### Key Entities *(include if feature involves data)*

- **Request Log Row**: A recent gateway request summary with status, latency, model, safe key label, route outcome, and provider-chain summary.
- **Request Detail**: Safe detail for one request, including redacted headers, trace, prompt preview, and response preview.
- **Provider Stage**: One provider attempt with provider identity, outcome, latency, and safe error summary.
- **Log Filter**: Status, fallback, pagination, and limit controls used to narrow request history.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can identify whether a recent seeded request succeeded, failed, or used fallback from the log list with 100% accuracy.
- **SC-002**: Operators can identify failed primary and successful fallback providers for a fallback request from its detail in one lookup.
- **SC-003**: Secret-like values are redacted in 100% of list and detail fixtures containing sensitive data.
- **SC-004**: Filtering by status group and fallback returns only matching rows for deterministic seeded data.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing proxy request behavior, streaming behavior, cost calculation, and audit logging remain unchanged for clients.
- **CC-002**: Request log capture does not require operators to query raw database tables.
- **CC-003**: Log query failures return compatible admin errors without exposing internal storage details.
- **CC-004**: Captured previews are bounded so operational visibility does not become a full prompt/response archive.

## Assumptions

- Request logs are for recent operational investigation, not long-term compliance retention.
- Audit history remains the immutable compliance record; request logs optimize incident investigation and trace display.
- Full raw prompt and response viewing is out of scope.
- Real-time log streaming is out of scope for the first backend capability.
