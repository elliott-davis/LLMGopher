# Research: UI API Key Lifecycle Controls

## Decision: Build on existing admin key APIs

**Rationale**: `internal/api/admin.go` already exposes `POST /v1/admin/keys`, `PUT /v1/admin/keys/{id}`, and `DELETE /v1/admin/keys/{id}` with support for `name`, `rate_limit_rps`, `expires_at`, `metadata`, `allowed_models`, and `is_active`. Reusing those endpoints keeps the UI aligned with spec 05 and avoids duplicating lifecycle policy outside the gateway.

**Alternatives considered**: Adding new UI-specific gateway endpoints was rejected because the existing admin contract already contains the needed fields. Direct database access from the UI was rejected because it bypasses validation, state cache behavior, and backend error handling.

## Decision: Keep lifecycle state server-authoritative

**Rationale**: The UI should fetch keys and models with `cache: "no-store"` and call `revalidatePath("/keys")` after mutations, matching the current create flow. This keeps the gateway state cache and database as the source of truth while accounting for the existing 5-second cache synchronization window.

**Alternatives considered**: Optimistic-only UI updates were rejected because key deletion, stale edits, and cache delay are explicit edge cases. Client-side persistence was rejected because it risks drift and secret exposure.

## Decision: Represent metadata as editable JSON object text

**Rationale**: The backend contract accepts `metadata` as a `map[string]string`. A JSON text area with client/server validation can preserve arbitrary string tags while giving operators clear validation errors without inventing a constrained tag UI prematurely.

**Alternatives considered**: A key/value row editor was considered for discoverability but would add more interaction complexity. Free-form unvalidated text was rejected because invalid metadata must be rejected without losing form state.

## Decision: Source model allowlist choices from existing model inventory

**Rationale**: `GET /v1/admin/models` already provides human-readable model information (`name`, `alias`, provider reference) and the exact model identifiers needed for enforcement. The UI can display friendly labels while submitting exact model IDs or aliases consistent with backend enforcement.

**Alternatives considered**: Manual comma-separated allowlists were rejected as the primary interaction because they are error-prone and fail FR-007. A manual fallback may still be useful for stale model edge cases only if preserving exact identifiers requires it.

## Decision: Use explicit confirmation for permanent deletion

**Rationale**: Deletion removes the API key and related budget row. The UI should require a confirmation dialog that names the key and explains that clients using the key will stop authenticating after gateway synchronization.

**Alternatives considered**: Immediate row-action deletion was rejected because FR-004 and SC-004 require explicit confirmation. Soft-delete only was rejected because the backend already distinguishes reversible `is_active` updates from permanent delete.

## Decision: Preserve raw key one-time display

**Rationale**: The backend only returns raw key material from create responses. The UI must continue showing the generated secret only in the create dialog and must not add raw key fields to edit, list, toast, or error states.

**Alternatives considered**: Adding a reveal or copy action for existing keys was rejected because plaintext keys are not persisted and recovering them would violate the security model.
