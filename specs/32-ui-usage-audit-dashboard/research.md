# Research: UI Usage and Audit Dashboard

## Decision: Treat Existing Admin Endpoints as Source of Truth

Use `GET /v1/admin/usage`, `GET /v1/admin/usage/daily`, and `GET /v1/admin/audit` as the UI contracts. Do not add new backend endpoints during initial implementation.

**Rationale**: The backend already exposes grouped usage summaries, daily usage summaries, and paginated audit records from PostgreSQL. The feature goal is UI/API parity, and reusing the existing handlers avoids drift from the tested storage/query behavior.

**Alternatives considered**: A new aggregate dashboard endpoint was considered, but it would duplicate existing query logic and make future backend changes harder to keep consistent.

## Decision: Fetch Analytics Server-Side with Admin Bearer Token

Add UI server-side fetch helpers that include `Authorization: Bearer ${LLMGOPHER_UI_ADMIN_API_KEY}` for analytics endpoints.

**Rationale**: The existing usage and audit routes are wrapped with admin auth. The UI already uses `LLMGOPHER_UI_ADMIN_API_KEY` for API key budgets, so the same server-only configuration pattern keeps credentials out of the browser and provides a clear unavailable state when the token is missing or rejected.

**Alternatives considered**: Client-side fetching was rejected because it would expose the admin token. Adding a separate UI auth flow is larger than this parity feature and belongs in a dedicated security/RBAC effort.

## Decision: Use Query Parameters for Filter State

Persist grouping, time window, API key, model, provider, status, limit, and offset in the page URL query string.

**Rationale**: URL-backed filters preserve selections across refreshes, server-rendered errors, pagination, and sharable admin investigations. This matches the backend query contract directly and avoids additional client state machinery.

**Alternatives considered**: Local component state would make pagination and refresh behavior less predictable. A persistent saved-filter model would require new storage and is out of scope.

## Decision: Render Daily Trends as Accessible Tables First

Display daily trend data in a table/cards format using existing UI primitives, with charting deferred until a chart dependency or design system pattern is chosen.

**Rationale**: The current UI has tables/cards but no chart dependency. Tables satisfy the acceptance criteria, are testable with existing React Testing Library patterns, and avoid adding a dependency for a single screen.

**Alternatives considered**: A chart library could improve visual scanning, but adds bundle weight, design decisions, and accessibility work not required for the first parity release.

## Decision: Keep Historical IDs Visible but Never Secrets

Show API key IDs from audit/usage rows as identifiers, not raw API keys or hashes beyond existing inventory behavior. Do not expose provider credentials or request payloads.

**Rationale**: Historical records can refer to deleted keys/models/providers, and administrators need stable identifiers for investigation. The backend audit contract does not include secrets or payload bodies, so the UI should not invent or fetch sensitive context.

**Alternatives considered**: Joining current key names or model/provider display names could improve readability, but stale/deleted entities make that best-effort. It can be added later without changing the analytics contract.

## Decision: Focus Tests on UI Contracts and State Handling

Use Vitest and React Testing Library for helper parsing, authenticated fetch behavior, page rendering, empty/unavailable states, invalid filters, and pagination preservation. Run Go tests only if backend files are changed.

**Rationale**: The implementation is planned as a UI parity feature over existing Go endpoints that already have handler and storage tests. Frontend behavior carries the new risk.

**Alternatives considered**: End-to-end browser tests would be useful but are not currently established in the repo. They can follow once a browser test harness exists.
