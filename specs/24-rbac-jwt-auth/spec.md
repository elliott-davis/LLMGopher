# Feature Specification: RBAC & JWT Authentication

**Feature Branch**: `[24-rbac-jwt-auth]`  
**Created**: 2026-04-26  
**Status**: Draft - pending implementation  
**Input**: Converted from `plans/24-rbac-jwt-auth.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Authorize Admin Users (Priority: P1)

Gateway administrators use JWT and RBAC policies for admin APIs.

**Why this priority**: Add role-based access control to the admin API and provide JWT as an alternative authentication method. This enables safe multi-operator deployments where different users have different levels of access.

**Independent Test**: Exercise the primary success path for RBAC & JWT Authentication through the gateway and verify the public behavior matches the converted acceptance criteria.

**Acceptance Scenarios**:

1. **Given** the gateway is configured for RBAC & JWT Authentication, **When** A proxy_viewer key can call GET /v1/admin/audit but receives 403 on POST /v1/admin/keys, **Then** the behavior is accepted for this feature.
2. **Given** the gateway is configured for RBAC & JWT Authentication, **When** An org_admin key can create teams within their org but not in another org, **Then** the behavior is accepted for this feature.
3. **Given** the gateway is configured for RBAC & JWT Authentication, **When** A key_owner key receives 403 on all admin endpoints except their own budget/usage, **Then** the behavior is accepted for this feature.

---

### User Story 2 - Handle Policy and Failure Cases (Priority: P2)

Operators and API clients receive predictable behavior when requests are invalid, providers fail, policy limits apply, or required configuration is absent.

**Why this priority**: Gateway features affect shared spend, availability, and provider credentials, so failure behavior must be explicit and safe.

**Independent Test**: Trigger the relevant negative paths from the converted plan and verify compatible errors, redaction, and audit context.

**Acceptance Scenarios**:

1. **Given** an invalid or unsupported request for RBAC & JWT Authentication, **When** the gateway rejects it, **Then** the response uses the expected public error contract.
2. **Given** required configuration is missing or disabled, **When** the feature is invoked, **Then** the gateway fails safely without leaking credentials or internal state.

---

### User Story 3 - Verify Operational Readiness (Priority: P3)

Gateway operators can test, observe, and roll out RBAC & JWT Authentication without losing request traceability or spend accountability.

**Why this priority**: A gateway feature is only production-ready when it can be verified, audited, and debugged under normal operating conditions.

**Independent Test**: Run the automated tests and quickstart smoke checks, then inspect logs, metrics, traces, or audit records relevant to this feature.

**Acceptance Scenarios**:

1. **Given** the feature is exercised successfully, **When** operators inspect runtime evidence, **Then** request IDs and audit or observability context are preserved.
2. **Given** the feature affects usage or provider calls, **When** a request completes, **Then** cost and audit behavior remains consistent with existing gateway rules.

### Edge Cases

- SSO / OIDC integration
- UI-based role management (update the Next.js UI in a separate task)
- Per-endpoint permission tables (role enum is sufficient)
- Token rotation

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST deliver RBAC & JWT Authentication for the gateway behavior described in the converted roadmap plan.
- **FR-002**: The system MUST preserve OpenAI-compatible request, response, and error semantics where the feature touches public APIs.
- **FR-003**: The system MUST preserve existing authentication, rate limiting, guardrail, audit, and cost behavior unless the feature explicitly changes those policies.
- **FR-004**: The system MUST provide tests for success, failure, and compatibility behavior described by the acceptance criteria.
- **FR-005**: The system MUST satisfy this acceptance behavior: A proxy_viewer key can call GET /v1/admin/audit but receives 403 on POST /v1/admin/keys.
- **FR-006**: The system MUST satisfy this acceptance behavior: An org_admin key can create teams within their org but not in another org.
- **FR-007**: The system MUST satisfy this acceptance behavior: A key_owner key receives 403 on all admin endpoints except their own budget/usage.
- **FR-008**: The system MUST satisfy this acceptance behavior: A valid JWT with role: "proxy_admin" claim has full admin access.
- **FR-009**: The system MUST satisfy this acceptance behavior: An expired JWT returns 401.
- **FR-010**: The system MUST satisfy this acceptance behavior: The master key always has full access regardless of role checks.

### Key Entities *(include if feature involves data)*

- **Gateway Feature Configuration**: Runtime settings, database records, or provider metadata that enable RBAC & JWT Authentication.
- **Client Request**: The user-facing request or admin operation that exercises the feature.
- **Gateway Decision**: The routing, policy, provider, cache, audit, or observability decision made by the gateway.
- **Verification Evidence**: Tests, smoke checks, audit records, metrics, or traces proving the behavior works.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The primary RBAC & JWT Authentication workflow succeeds through the gateway using the public API or admin API described by the converted plan.
- **SC-002**: Negative and unsupported cases return predictable compatible errors without leaking secrets.
- **SC-003**: Automated tests cover the converted acceptance criteria for success and failure paths.
- **SC-004**: Functional smoke verification is recorded before the feature is marked verified.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing clients that do not use RBAC & JWT Authentication continue to work without request or response contract changes.
- **CC-002**: Request IDs, structured logs, audit context, and redaction are preserved for runtime paths touched by the feature.
- **CC-003**: Any provider call, budget change, routing decision, cache event, or observability event introduced by the feature remains testable and auditable.
- **CC-004**: Configuration follows the established precedence and startup validation rules.

## Assumptions

- The original plan marks this feature pending; implementation and verification are both required.
- Converted dependencies: none beyond existing gateway infrastructure.
- Out-of-scope boundaries from the original plan remain unchanged: SSO / OIDC integration; UI-based role management (update the Next.js UI in a separate task); Per-endpoint permission tables (role enum is sufficient)
- The detailed implementation notes from the Claude plan are preserved in `plan.md`, while this spec focuses on user-visible and operational outcomes.
