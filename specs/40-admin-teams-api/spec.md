# Feature Specification: Admin Teams API

**Feature Branch**: `[40-admin-teams-api]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User request to define backend capabilities that support the new admin Teams UI without creating one mega-spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Team Governance Summary (Priority: P1)

Administrators view teams with member counts and budget health so they can understand tenant ownership and spend posture from the admin UI.

**Why this priority**: The Teams UI requires real tenant data and governance state; current backend support does not expose a production teams admin contract.

**Independent Test**: Seed teams with different member counts and budget utilization states; query the team list; verify rows include safe identity, member counts, and budget health.

**Acceptance Scenarios**:

1. **Given** teams exist, **When** an administrator lists teams, **Then** each team includes a stable id, display name, member count, budget utilization, and budget health state.
2. **Given** a team is at or above its budget alert threshold, **When** the team is returned, **Then** the response marks the team as near cap or over cap.
3. **Given** a team has no configured budget, **When** teams are listed, **Then** the response clearly distinguishes missing budget policy from healthy usage.
4. **Given** no teams exist, **When** teams are listed, **Then** the response distinguishes an empty organization from an unavailable teams service.

---

### User Story 2 - Preserve Team Reference Integrity (Priority: P2)

Operators can understand request, budget, and rate-limit records that reference teams even when related keys or members change over time.

**Why this priority**: Teams become shared references for budgets, rate limits, audit records, and future RBAC features, so stable identity matters before mutation-heavy UI is added.

**Independent Test**: Create team-linked budget and rate-limit records, change membership or remove a key reference, and verify team summaries remain stable and safe.

**Acceptance Scenarios**:

1. **Given** a team has associated keys, budgets, or rate-limit rules, **When** the team summary is returned, **Then** it includes safe counts or links without exposing raw key material.
2. **Given** a referenced key is revoked or deleted, **When** the team summary is returned, **Then** the team still appears with a missing-reference indication where needed.
3. **Given** an administrator lacks future team mutation permission, **When** teams are queried, **Then** read access follows admin authorization rules and does not expose unauthorized mutation affordances.

### Edge Cases

- A team has zero members.
- A team has active keys but no budget.
- A team is over budget but has no current active members.
- Team names collide after normalization.
- A team referenced by historical records is deleted or archived.
- Member identity details are unavailable from the authentication system.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose a list of teams with stable id, display name, member count, budget utilization, and budget health state.
- **FR-002**: The system MUST mark teams at or above alert threshold and teams over cap using explicit state values.
- **FR-003**: The system MUST distinguish no teams, no budget policy, and teams service unavailable states.
- **FR-004**: The system MUST avoid returning raw API keys, credentials, or unnecessary member personal data in team summaries.
- **FR-005**: The system MUST maintain stable team references for budgets, rate limits, audit records, and future RBAC integration.
- **FR-006**: The system MUST authenticate and authorize team reads consistently with admin endpoint rules.
- **FR-007**: The system MUST document or reject team creation, deletion, renaming, membership, and role-assignment mutations until those operations are explicitly supported.
- **FR-008**: The system MUST preserve OpenAI-compatible gateway behavior and avoid changing request proxy semantics.
- **FR-009**: The system MUST provide tests for populated teams, warning state, no teams, no budget policy, missing references, and unavailable service behavior.

### Key Entities *(include if feature involves data)*

- **Team**: A tenant or organizational grouping with stable identity, display name, membership count, and governance metadata.
- **Team Budget Health**: Derived state showing whether a team is healthy, near cap, over cap, or missing budget policy.
- **Team Reference**: A stable link from keys, budgets, rate limits, audit records, or future RBAC assignments to a team.
- **Member Summary**: A count or safe aggregate of team membership that avoids unnecessary personal data exposure.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can see every seeded team with display name and member count from one team list query.
- **SC-002**: Teams at or above configured budget thresholds are classified correctly in 100% of budget health fixtures.
- **SC-003**: Team summaries expose zero raw API keys or provider credentials across all fixtures.
- **SC-004**: Empty organization, missing budget, and unavailable service states are distinguishable in response contracts.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing API key, budget, audit, and rate-limit behavior remains unchanged unless explicitly tied to team governance by later specs.
- **CC-002**: Team references remain stable enough for audit and governance history.
- **CC-003**: Operators can inspect team governance state without direct database access.

## Assumptions

- The first Teams backend capability is read-only.
- Team membership management and role assignment remain out of scope for this spec.
- Future RBAC integration may add actor-specific permissions without changing the team summary requirements.
- Budget health may be derived from budget data owned by the budgets feature.
