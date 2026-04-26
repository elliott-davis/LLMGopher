# Feature Specification: Budget Lifecycle (Alerts & Periodic Resets)

**Feature Branch**: `[10-budget-lifecycle]`  
**Created**: 2026-04-26  
**Status**: Draft - implemented, functional verification needed  
**Input**: Converted from `plans/10-budget-lifecycle.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automate Budget Lifecycle (Priority: P1)

Gateway operators receive budget alerts and periodic budget resets without manual database changes.

**Why this priority**: Add two missing budget features: (1) soft alerts that warn when a key approaches its limit, and (2) automatic periodic budget resets so keys don't have to be manually reset each billing cycle.

**Independent Test**: Exercise the primary success path for Budget Lifecycle (Alerts & Periodic Resets) through the gateway and verify the public behavior matches the converted acceptance criteria.

**Acceptance Scenarios**:

1. **Given** the gateway is configured for Budget Lifecycle (Alerts & Periodic Resets), **When** A key at 80% of its budget with alert_threshold_pct: 80 logs a budget_threshold_reached warning, **Then** the behavior is accepted for this feature.
2. **Given** the gateway is configured for Budget Lifecycle (Alerts & Periodic Resets), **When** The alert does not fire again within 1 hour for the same key, **Then** the behavior is accepted for this feature.
3. **Given** the gateway is configured for Budget Lifecycle (Alerts & Periodic Resets), **When** A key with budget_duration: "monthly" and budget_reset_at in the past gets spent_usd reset to 0, **Then** the behavior is accepted for this feature.

---

### User Story 2 - Handle Policy and Failure Cases (Priority: P2)

Operators and API clients receive predictable behavior when requests are invalid, providers fail, policy limits apply, or required configuration is absent.

**Why this priority**: Gateway features affect shared spend, availability, and provider credentials, so failure behavior must be explicit and safe.

**Independent Test**: Trigger the relevant negative paths from the converted plan and verify compatible errors, redaction, and audit context.

**Acceptance Scenarios**:

1. **Given** an invalid or unsupported request for Budget Lifecycle (Alerts & Periodic Resets), **When** the gateway rejects it, **Then** the response uses the expected public error contract.
2. **Given** required configuration is missing or disabled, **When** the feature is invoked, **Then** the gateway fails safely without leaking credentials or internal state.

---

### User Story 3 - Verify Operational Readiness (Priority: P3)

Gateway operators can test, observe, and roll out Budget Lifecycle (Alerts & Periodic Resets) without losing request traceability or spend accountability.

**Why this priority**: A gateway feature is only production-ready when it can be verified, audited, and debugged under normal operating conditions.

**Independent Test**: Run the automated tests and quickstart smoke checks, then inspect logs, metrics, traces, or audit records relevant to this feature.

**Acceptance Scenarios**:

1. **Given** the feature is exercised successfully, **When** operators inspect runtime evidence, **Then** request IDs and audit or observability context are preserved.
2. **Given** the feature affects usage or provider calls, **When** a request completes, **Then** cost and audit behavior remains consistent with existing gateway rules.

### Edge Cases

- Email or webhook alerts (covered later in spec 22)
- Per-model budgets (spec 23)
- Retroactive reset (only resets at scheduled budget_reset_at, not backdated)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST deliver Budget Lifecycle (Alerts & Periodic Resets) for the gateway behavior described in the converted roadmap plan.
- **FR-002**: The system MUST preserve OpenAI-compatible request, response, and error semantics where the feature touches public APIs.
- **FR-003**: The system MUST preserve existing authentication, rate limiting, guardrail, audit, and cost behavior unless the feature explicitly changes those policies.
- **FR-004**: The system MUST provide tests for success, failure, and compatibility behavior described by the acceptance criteria.
- **FR-005**: The system MUST satisfy this acceptance behavior: A key at 80% of its budget with alert_threshold_pct: 80 logs a budget_threshold_reached warning.
- **FR-006**: The system MUST satisfy this acceptance behavior: The alert does not fire again within 1 hour for the same key.
- **FR-007**: The system MUST satisfy this acceptance behavior: A key with budget_duration: "monthly" and budget_reset_at in the past gets spent_usd reset to 0.
- **FR-008**: The system MUST satisfy this acceptance behavior: budget_reset_at advances by one month after reset.
- **FR-009**: The system MUST satisfy this acceptance behavior: A key without budget_duration is never auto-reset.
- **FR-010**: The system MUST satisfy this acceptance behavior: Migration runs cleanly on existing schema.

### Key Entities *(include if feature involves data)*

- **Gateway Feature Configuration**: Runtime settings, database records, or provider metadata that enable Budget Lifecycle (Alerts & Periodic Resets).
- **Client Request**: The user-facing request or admin operation that exercises the feature.
- **Gateway Decision**: The routing, policy, provider, cache, audit, or observability decision made by the gateway.
- **Verification Evidence**: Tests, smoke checks, audit records, metrics, or traces proving the behavior works.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The primary Budget Lifecycle (Alerts & Periodic Resets) workflow succeeds through the gateway using the public API or admin API described by the converted plan.
- **SC-002**: Negative and unsupported cases return predictable compatible errors without leaking secrets.
- **SC-003**: Automated tests cover the converted acceptance criteria for success and failure paths.
- **SC-004**: Functional smoke verification is recorded before the feature is marked verified.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing clients that do not use Budget Lifecycle (Alerts & Periodic Resets) continue to work without request or response contract changes.
- **CC-002**: Request IDs, structured logs, audit context, and redaction are preserved for runtime paths touched by the feature.
- **CC-003**: Any provider call, budget change, routing decision, cache event, or observability event introduced by the feature remains testable and auditable.
- **CC-004**: Configuration follows the established precedence and startup validation rules.

## Assumptions

- The original plan marks this feature complete; functional smoke verification is still required.
- Converted dependencies: none beyond existing gateway infrastructure.
- Out-of-scope boundaries from the original plan remain unchanged: Email or webhook alerts (covered later in spec 22); Per-model budgets (spec 23); Retroactive reset (only resets at scheduled budget_reset_at, not backdated)
- The detailed implementation notes from the Claude plan are preserved in `plan.md`, while this spec focuses on user-visible and operational outcomes.
