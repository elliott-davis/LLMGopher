# Feature Specification: Teams & Organizations

**Feature Branch**: `[23-teams-organizations]`  
**Created**: 2026-04-26  
**Status**: Draft - pending implementation  
**Input**: Converted from `plans/23-teams-organizations.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage Teams and Organizations (Priority: P1)

Gateway administrators organize keys, users, budgets, and usage by tenant.

**Why this priority**: Add a team and organization model so that API keys can be grouped, team-level budgets and rate limits can be enforced, and usage can be attributed at the team level. This is the foundation for multi-tenant operation.

**Independent Test**: Exercise the primary success path for Teams & Organizations through the gateway and verify the public behavior matches the converted acceptance criteria.

**Acceptance Scenarios**:

1. **Given** the gateway is configured for Teams & Organizations, **When** Organization and team can be created via admin API, **Then** the behavior is accepted for this feature.
2. **Given** the gateway is configured for Teams & Organizations, **When** An API key can be assigned to a team, **Then** the behavior is accepted for this feature.
3. **Given** the gateway is configured for Teams & Organizations, **When** A team with an exhausted budget returns 429 on the next request, **Then** the behavior is accepted for this feature.

---

### User Story 2 - Handle Policy and Failure Cases (Priority: P2)

Operators and API clients receive predictable behavior when requests are invalid, providers fail, policy limits apply, or required configuration is absent.

**Why this priority**: Gateway features affect shared spend, availability, and provider credentials, so failure behavior must be explicit and safe.

**Independent Test**: Trigger the relevant negative paths from the converted plan and verify compatible errors, redaction, and audit context.

**Acceptance Scenarios**:

1. **Given** an invalid or unsupported request for Teams & Organizations, **When** the gateway rejects it, **Then** the response uses the expected public error contract.
2. **Given** required configuration is missing or disabled, **When** the feature is invoked, **Then** the gateway fails safely without leaking credentials or internal state.

---

### User Story 3 - Verify Operational Readiness (Priority: P3)

Gateway operators can test, observe, and roll out Teams & Organizations without losing request traceability or spend accountability.

**Why this priority**: A gateway feature is only production-ready when it can be verified, audited, and debugged under normal operating conditions.

**Independent Test**: Run the automated tests and quickstart smoke checks, then inspect logs, metrics, traces, or audit records relevant to this feature.

**Acceptance Scenarios**:

1. **Given** the feature is exercised successfully, **When** operators inspect runtime evidence, **Then** request IDs and audit or observability context are preserved.
2. **Given** the feature affects usage or provider calls, **When** a request completes, **Then** cost and audit behavior remains consistent with existing gateway rules.

### Edge Cases

- Team member management (users within teams - that's part of RBAC in spec 24)
- Per-team model access lists (add after spec 24 RBAC is in place)
- Team-level rate limits (the rate_limit_team_rps column is reserved for a future spec)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST deliver Teams & Organizations for the gateway behavior described in the converted roadmap plan.
- **FR-002**: The system MUST preserve OpenAI-compatible request, response, and error semantics where the feature touches public APIs.
- **FR-003**: The system MUST preserve existing authentication, rate limiting, guardrail, audit, and cost behavior unless the feature explicitly changes those policies.
- **FR-004**: The system MUST provide tests for success, failure, and compatibility behavior described by the acceptance criteria.
- **FR-005**: The system MUST satisfy this acceptance behavior: Organization and team can be created via admin API.
- **FR-006**: The system MUST satisfy this acceptance behavior: An API key can be assigned to a team.
- **FR-007**: The system MUST satisfy this acceptance behavior: A team with an exhausted budget returns 429 on the next request.
- **FR-008**: The system MUST satisfy this acceptance behavior: Team budget deduction is atomic (concurrent requests don't over-spend).
- **FR-009**: The system MUST satisfy this acceptance behavior: GET /v1/admin/usage?group_by=team returns usage broken down by team.
- **FR-010**: The system MUST satisfy this acceptance behavior: Keys without a team_id continue to work unchanged.

### Key Entities *(include if feature involves data)*

- **Gateway Feature Configuration**: Runtime settings, database records, or provider metadata that enable Teams & Organizations.
- **Client Request**: The user-facing request or admin operation that exercises the feature.
- **Gateway Decision**: The routing, policy, provider, cache, audit, or observability decision made by the gateway.
- **Verification Evidence**: Tests, smoke checks, audit records, metrics, or traces proving the behavior works.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The primary Teams & Organizations workflow succeeds through the gateway using the public API or admin API described by the converted plan.
- **SC-002**: Negative and unsupported cases return predictable compatible errors without leaking secrets.
- **SC-003**: Automated tests cover the converted acceptance criteria for success and failure paths.
- **SC-004**: Functional smoke verification is recorded before the feature is marked verified.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing clients that do not use Teams & Organizations continue to work without request or response contract changes.
- **CC-002**: Request IDs, structured logs, audit context, and redaction are preserved for runtime paths touched by the feature.
- **CC-003**: Any provider call, budget change, routing decision, cache event, or observability event introduced by the feature remains testable and auditable.
- **CC-004**: Configuration follows the established precedence and startup validation rules.

## Assumptions

- The original plan marks this feature pending; implementation and verification are both required.
- Converted dependencies: none beyond existing gateway infrastructure.
- Out-of-scope boundaries from the original plan remain unchanged: Team member management (users within teams - that's part of RBAC in spec 24); Per-team model access lists (add after spec 24 RBAC is in place); Team-level rate limits (the rate_limit_team_rps column is reserved for a future spec)
- The detailed implementation notes from the Claude plan are preserved in `plan.md`, while this spec focuses on user-visible and operational outcomes.
