# Research: UI API Key Budget Controls

## Decision: Use existing spec 07 budget endpoints as the source of truth

**Rationale**: The gateway already exposes key-scoped budget operations:

- `GET /v1/admin/keys/{id}/budget`
- `PUT /v1/admin/keys/{id}/budget`
- `DELETE /v1/admin/keys/{id}/budget`
- `POST /v1/admin/keys/{id}/budget/reset`

These routes map directly to the feature's view, set/update, remove, and reset scenarios and already have Go handler and storage tests.

**Alternatives considered**:

- Add budget fields to `GET /v1/admin/keys`: rejected for this feature because it expands the existing key listing API and is not necessary for parity with the existing budget API.
- Add a new UI-only aggregation endpoint: rejected because it creates another contract to maintain before there is evidence the existing endpoint shape is insufficient.

## Decision: Treat missing budget as a first-class UI state

**Rationale**: A key without a budget currently returns `404` from `GET /v1/admin/keys/{id}/budget`. The UI should map that specific response to "No budget set" with a setup action, not to a generic failure state.

**Alternatives considered**:

- Change the backend to return `200` with nullable budget data: rejected to avoid changing the established spec 07 API behavior.
- Hide budget controls until a budget exists: rejected because administrators need to create budgets from the UI.

## Decision: Keep budget controls inside the API key management surface

**Rationale**: Budgets are scoped to one API key. The existing `/keys` page already lists key inventory and has per-row actions for edit, activate/deactivate, and delete. Adding a budget status indicator plus a manage-budget action keeps the operational workflow localized.

**Alternatives considered**:

- Add a standalone `/budgets` route: rejected for initial delivery because it would duplicate key identity context and broaden navigation scope.
- Put budget fields in the existing edit-key form: rejected because budget lifecycle operations include reset and remove confirmations that are distinct from key metadata changes.

## Decision: Use a server-only UI admin API key for protected budget calls

**Rationale**: Budget routes are protected by `applyAuthMiddleware`, while existing key list and key mutation routes are not. The UI must call budget routes from server actions/server components with an `Authorization: Bearer ...` header sourced from server-only configuration. A missing token should produce a clear unavailable/authorization setup message rather than exposing auth details to the browser.

**Alternatives considered**:

- Send a gateway API key from the browser: rejected because it exposes administrative credentials to client components.
- Remove auth from budget routes: rejected because budget controls alter spend governance and should remain protected.
- Block planning until a full admin session model exists: rejected because the spec scopes to trusted gateway administrators and the existing Compose dev key can support local operation.

## Decision: Mirror gateway validation in UI helpers, then rely on gateway errors for final authority

**Rationale**: The gateway requires `budget_usd > 0`, `alert_threshold_pct` between 1 and 99 when set, valid duration values, and `budget_reset_at` when duration is set. Client/UI-side validation improves feedback, but the gateway remains authoritative for race conditions and cross-admin updates.

**Alternatives considered**:

- Let the gateway handle all validation: rejected because the spec requires clear feedback before or during save and UI validation is cheap for deterministic fields.
- Add broader validation rules not present in the gateway: rejected because it could create UI/API drift.

## Decision: Use focused polling/refresh after mutations

**Rationale**: Existing UI actions call `revalidatePath("/keys")`, `router.refresh()`, and short delayed refreshes for cache synchronization. Budget actions should follow that pattern and fetch the latest budget state after set/update/reset/delete.

**Alternatives considered**:

- Add real-time subscriptions: rejected as unnecessary for admin budget lifecycle controls.
- Optimistically update all state without refetch: rejected because spend may change asynchronously and reset/delete can race with other administrators.
