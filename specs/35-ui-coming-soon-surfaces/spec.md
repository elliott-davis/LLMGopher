# Feature Specification: UI Coming Soon Surfaces

**Feature Branch**: `[35-ui-coming-soon-surfaces]`  
**Created**: 2026-05-09  
**Status**: Draft  
**Input**: User description: "Look at the UI mocks and write specs for the missing functionality gaps. Specifically the pages that say 'coming soon' need well defined specs."

## Context

The admin UI sidebar exposes eight routes whose pages currently render only `Coming soon.`: Logs, Audit, Routes, Guardrails, Teams, Budgets, Rate Limits, and Settings. The E2E mock backend already defines seed data and contracts for several of these surfaces, while existing product specs cover the underlying backend capabilities for teams, budgets, guardrails, fallback, load balancing, per-model rate limits, and audit.

This specification turns those placeholders into implementation-ready UI scope. Each route has a focused page spec under `surface-specs/`:

- `logs.md`
- `audit.md`
- `routes.md`
- `guardrails.md`
- `teams.md`
- `budgets.md`
- `rate-limits.md`
- `settings.md`

## Clarifications

### Session 2026-05-09

- Q: How should mutation behavior be handled before each production admin API is reconciled? → A: Production pages are read-only/unavailable until real contracts exist; mock-backed E2E may still cover intended UI states.
- Q: How should prompt and response bodies be displayed in request inspectors? → A: Show redacted, truncated previews only; full raw prompt/response viewing is out of scope.
- Q: How should Logs and Audit filter state persist across refresh or sharing? → A: Logs and Audit filters use URL query parameters and survive browser reload/share.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Investigate Gateway Traffic (Priority: P1)

Gateway operators use Logs and Audit pages to understand recent request behavior, routing outcomes, failures, spend, and administrative history without querying raw storage or calling admin APIs manually.

**Why this priority**: Request and audit visibility are the highest-value operational gap among the placeholders because they support incident response, spend investigation, and compliance review.

**Independent Test**: Load `/logs` and `/audit` against the deterministic mock backend, filter the seeded rows, open a fallback log detail, and verify no secret-bearing headers or credentials are exposed.

**Acceptance Scenarios**:

1. **Given** mixed successful, client-error, server-error, and fallback request rows, **When** an operator opens Logs, **Then** the page shows a filterable request table with status, latency, key, model, and routing outcome.
2. **Given** a fallback-handled request, **When** an operator opens its inspector, **Then** the trace identifies the failed primary provider and successful fallback stage.
3. **Given** audit records are available, **When** an operator filters by actor, action, or date range, **Then** the Audit page narrows results while preserving a paginated, newest-first history.
4. **Given** log detail includes headers, prompts, and responses, **When** the inspector renders, **Then** sensitive headers and credentials are redacted and prompt/response bodies appear only as truncated previews.

---

### User Story 2 - Configure Routing and Safety Controls (Priority: P1)

Gateway administrators configure model routing strategies, fallback behavior, and guardrail toggles from the UI instead of editing database state or issuing raw admin API calls.

**Why this priority**: Routing and safety controls directly affect request reliability and policy enforcement for all clients.

**Independent Test**: Load `/routes` and `/guardrails` against seeded providers/models/guardrails, switch strategy views, toggle `gr_jail`, reload, and confirm the persisted guardrail state is retained.

**Acceptance Scenarios**:

1. **Given** configured providers and models, **When** an administrator opens Routes, **Then** they can inspect each route's strategy, provider order, fallback path, and health impact.
2. **Given** a route supports weighted, fallback, latency, and single-provider strategies, **When** the administrator changes the selected strategy in the UI, **Then** the route diagram and editable controls reflect the selected strategy clearly.
3. **Given** a reconciled guardrail mutation contract or mock-backed E2E fixture, **When** an administrator enables a disabled guardrail and reloads the page, **Then** the Guardrails page shows the enabled state.
4. **Given** a reconciled guardrail mutation contract or mock-backed E2E fixture, **When** a guardrail save fails, **Then** the UI restores or preserves state clearly and displays the failure reason.

---

### User Story 3 - Govern Tenants, Budgets, and Rate Limits (Priority: P1)

Gateway administrators use Teams, Budgets, and Rate Limits pages to monitor tenant-level usage controls, identify near-cap tenants, and manage throttling rules.

**Why this priority**: Multi-tenant spend and throttling governance are core LLMGopher admin responsibilities and are already represented in mocks and backend specs.

**Independent Test**: Load `/teams`, `/budgets`, and `/rate-limits` with seeded fixtures; verify both teams render, Research shows an 85%+ budget warning, and exactly one rate-limit rule shows a tripped state.

**Acceptance Scenarios**:

1. **Given** teams exist, **When** an administrator opens Teams, **Then** they see each team's name, member count, and budget utilization state.
2. **Given** a team is above its alert threshold, **When** Budgets renders, **Then** the team is visually and textually marked as near cap.
3. **Given** budget settings are editable through a reconciled mutation contract or mock-backed E2E fixture, **When** an administrator changes a limit, duration, or alert threshold, **Then** the saved value is reflected without losing existing spend state.
4. **Given** a rate-limit rule is tripped, **When** Rate Limits renders, **Then** the rule has a clear tripped indicator and enough scope context for an operator to identify the affected key, model, or team.

---

### User Story 4 - Manage Organization-Level Settings (Priority: P2)

Gateway administrators use Settings to review and update organization-level controls such as gateway identity, authentication posture, notifications, and UI preferences.

**Why this priority**: Settings is less urgent than operational and governance pages, but it needs a defined scope so the placeholder does not accumulate unrelated concerns.

**Independent Test**: Load `/settings` and verify all four settings cards render with clear save states, validation, and disabled/unavailable behavior for backend-dependent controls.

**Acceptance Scenarios**:

1. **Given** organization settings exist, **When** an administrator opens Settings, **Then** the page shows Gateway Profile, Security, Notifications, and Display cards.
2. **Given** a setting can be changed safely through a reconciled mutation contract or local-only display preference, **When** the administrator edits and saves it, **Then** the page confirms success and displays the persisted value.
3. **Given** a setting is not yet backed by an API, **When** the page renders, **Then** it is clearly marked unavailable rather than silently editable.

### Edge Cases

- The backing admin endpoint for a placeholder page has not shipped yet.
- The mock contract differs from the eventual gateway contract.
- A filtered list returns no rows.
- A paginated list has more rows than the first page limit.
- A mutation succeeds but a follow-up refresh fails.
- A request or audit record contains secret-like values in headers, prompts, responses, errors, or metadata.
- A referenced provider, model, key, team, guardrail, budget, or rate-limit scope has been deleted since the row was recorded.
- An administrator lacks permission for a destructive or policy-changing action once RBAC is enabled.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST replace every `Coming soon.` admin route with a page that has a defined primary user goal, visible loading state, empty state, unavailable/error state, and accessible page title.
- **FR-002**: The system MUST implement the Logs page according to `surface-specs/logs.md`.
- **FR-003**: The system MUST implement the Audit page according to `surface-specs/audit.md`.
- **FR-004**: The system MUST implement the Routes page according to `surface-specs/routes.md`.
- **FR-005**: The system MUST implement the Guardrails page according to `surface-specs/guardrails.md`.
- **FR-006**: The system MUST implement the Teams page according to `surface-specs/teams.md`.
- **FR-007**: The system MUST implement the Budgets page according to `surface-specs/budgets.md`.
- **FR-008**: The system MUST implement the Rate Limits page according to `surface-specs/rate-limits.md`.
- **FR-009**: The system MUST implement the Settings page according to `surface-specs/settings.md`.
- **FR-010**: The system MUST preserve OpenAI-compatible gateway behavior and avoid request hot-path changes while adding these admin UI surfaces.
- **FR-011**: The system MUST use typed UI contracts for every backend payload displayed or mutated by a placeholder replacement page.
- **FR-012**: The system MUST redact credentials, authorization headers, raw API keys, provider tokens, and other secret-like values from all table, detail, drawer, and error displays, and MUST show prompt/response bodies only as redacted, truncated previews.
- **FR-013**: The system MUST keep production-facing mutation controls read-only, disabled, or unavailable until the relevant real gateway contract has been reconciled; mock-backed E2E tests MAY exercise intended mutation states without enabling production writes.
- **FR-014**: The system MUST expose deterministic selectors only for interactions required by E2E coverage, favoring accessible roles and labels elsewhere.
- **FR-015**: The system MUST document any requested page capability that remains API-only or unavailable with a clear user-role rationale and follow-up trigger.
- **FR-016**: The system MUST encode Logs and Audit filter state in URL query parameters so filtered investigation views survive reloads and can be shared.

### Key Entities *(include if feature involves data)*

- **Request Log Row**: A recent gateway request with status, latency, model, API key, provider chain, and routing outcome.
- **Request Inspector Detail**: Request context for one log row, including trace, redacted prompt/response previews, and redacted headers.
- **Audit Record**: Immutable record of request or administrative activity, including actor-like context, action/filter fields, timestamps, outcome, cost, and error context.
- **Route Policy**: Model/provider routing configuration, including strategy, provider order or weights, fallback behavior, and health context.
- **Guardrail Rule**: Safety policy with enabled state and operator-facing identity.
- **Team**: Tenant grouping with display name, member count, and governance metadata.
- **Budget Policy**: Spend limit and alert threshold scoped to a team or key.
- **Rate Limit Rule**: Throttling rule scoped to key, model, or team, with request/token limits and tripped state.
- **Organization Setting**: Admin-editable or read-only setting grouped into Gateway Profile, Security, Notifications, or Display.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can identify whether a recent request succeeded, failed, or fell back in under 30 seconds from opening Logs.
- **SC-002**: Operators can open a fallback request and identify the failed primary provider plus successful fallback provider in under 45 seconds.
- **SC-003**: Administrators can identify teams above their configured budget alert threshold in under 30 seconds.
- **SC-004**: Administrators can identify every tripped rate-limit rule in under 30 seconds.
- **SC-005**: Administrators can toggle an existing guardrail and confirm persisted state after reload with one page interaction sequence.
- **SC-006**: All eight replacement pages pass the existing accessibility E2E gate with zero axe violations.
- **SC-007**: Secret-bearing values are redacted in 100% of rendered request detail and error states covered by fixtures.

### Compatibility & Operational Criteria *(include when relevant)*

- **CC-001**: Existing OpenAI-compatible client request/response behavior remains unchanged.
- **CC-002**: Admin UI additions do not add provider calls or gateway request hot-path work.
- **CC-003**: Mock-backed tests for these pages remain deterministic across reloads and test runs.
- **CC-004**: Surfaces that depend on unshipped real admin APIs remain clearly marked as mock/spec-backed until backend contracts are reconciled.
- **CC-005**: Operators can view or manage each placeholder surface through the admin UI without issuing raw API calls once its backing contract exists.

## Assumptions

- The target users are trusted gateway administrators and operators.
- The current Next.js admin shell and sidebar remain the navigation entry points for all eight pages.
- Existing backend feature specs remain the source of truth for runtime behavior: audit/usage APIs, budget lifecycle, teams, guardrails, provider fallback, load balancing, and per-model rate limits.
- The E2E mock backend fixtures under `ui/tests/fixtures/` represent the intended minimum UI data for first implementation.
- RBAC-aware action hiding is handled by the existing RBAC feature when roles ship; these page specs define the privileged administrator experience first.
- Mobile-specific layouts are out of scope for the first implementation, but pages must remain usable at the dashboard's supported responsive widths.
