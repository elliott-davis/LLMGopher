# Phase 0 Research: UI Coming Soon Surfaces

## Decision: Scope implementation to `ui/` only

**Rationale**: The feature replaces visible admin placeholders. Existing backend and mock specs already define the operational capabilities, and the feature requirement explicitly says production mutation controls must stay read-only/unavailable until real contracts are reconciled. Keeping this to `ui/` avoids changing gateway hot paths, provider routing, storage, audit workers, or config behavior.

**Alternatives considered**:

- Add missing backend endpoints first. Rejected for this planning slice because several pages already have mock contracts and the spec is about defining/replacing UI gaps.
- Build a generic admin surface framework. Rejected because the eight pages have distinct operator goals and acceptance criteria.

## Decision: Use typed TypeScript contracts at the UI boundary

**Rationale**: The constitution requires typed contracts and the spec requires typed UI contracts for every displayed or mutated payload. Existing mock handlers and fixtures already describe most shapes, so implementation should centralize surface-specific interfaces in `ui/src/lib/admin-surface-contracts.ts` or nearby feature modules and convert raw fetch responses before rendering.

**Alternatives considered**:

- Render raw JSON directly from admin endpoints. Rejected because it would leak backend drift into UI and increase redaction risk.
- Generate TypeScript from OpenAPI now. Rejected for this phase because the relevant production admin contracts are not uniformly available; generation should be revisited when those contracts stabilize.

## Decision: Keep production writes gated by contract readiness

**Rationale**: Guardrails, Budgets, Rate Limits, Routes, and Settings include intended mutation states, but the spec clarifies production pages must be read-only/unavailable until real gateway contracts exist. Mock-backed E2E may still exercise intended state transitions so the UX can be reviewed before backend reconciliation.

**Alternatives considered**:

- Enable writes against mock-only endpoints in production code. Rejected because it would create false affordances for administrators.
- Hide all controls completely. Rejected because disabled/unavailable states are part of the acceptance criteria and guide follow-up API work.

## Decision: Share redaction and truncation helpers across Logs and Audit

**Rationale**: Logs and Audit both display request-adjacent data that may contain authorization headers, provider credentials, raw API keys, prompts, responses, errors, and metadata. A shared helper makes redaction testable and prevents one page from drifting into less-safe behavior.

**Alternatives considered**:

- Redact separately in each page component. Rejected because duplicated secret handling is error-prone.
- Depend only on backend redaction. Rejected because the UI must never render secret-like values from fixtures, mocks, or future endpoints even if a backend response regresses.

## Decision: Persist Logs and Audit filters in URL query parameters

**Rationale**: The spec requires filtered investigation views to survive reload/share. A small shared query-state utility keeps parsing, defaults, and clear-filter behavior consistent across the two pages.

**Alternatives considered**:

- Store filters in React state only. Rejected because reload/share would lose investigation context.
- Store filters in local storage. Rejected because sharing filtered URLs is a stated requirement and local storage can leak stale filters between unrelated investigations.

## Decision: Prefer accessible roles and labels over broad `data-testid` usage

**Rationale**: Existing E2E guidance favors accessible selectors and deterministic `data-testid`s only where fixture-specific assertions require them. This supports both Playwright stability and accessibility requirements.

**Alternatives considered**:

- Add `data-testid` to every UI element. Rejected because it makes tests less user-centered and adds implementation noise.
- Use text-only queries everywhere. Rejected because some fixture-specific rows and diagrams need stable identifiers.

## Decision: Use existing mock fixtures as the first deterministic data source

**Rationale**: `ui/tests/fixtures/` already has seed data for logs, audit, teams, budgets, rate limits, and guardrails, plus mock handlers for the same groups. Implementing against those fixtures gives deterministic E2E coverage while production API reconciliation proceeds.

**Alternatives considered**:

- Run E2E against a live gateway and database. Rejected because the tests would be slower and less deterministic.
- Hand-code page-local fixtures. Rejected because it would drift from the E2E mock backend and existing test plan.
