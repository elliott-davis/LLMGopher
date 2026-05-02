# Data Model: UI API Key Lifecycle Controls

## API Key

Represents a managed gateway credential returned by `GET /v1/admin/keys` and mutated through the admin key endpoints.

**Fields**

- `id`: Stable UUID used for update, delete, budget, and row action targets.
- `key_hash`: SHA-256 hash prefix/value displayed for identification; never reversible to raw key material.
- `name`: Operator-friendly label. Required and non-empty.
- `rate_limit_rps`: Non-negative key-level request rate limit.
- `is_active`: Boolean lifecycle state. `false` disables authentication without deleting the record.
- `expires_at`: Optional timestamp after which the key is invalid.
- `metadata`: Optional string map for operational tags and attribution.
- `allowed_models`: Optional list of exact model identifiers allowed for this key. Empty means unrestricted.
- `created_at`: Creation timestamp.
- `updated_at`: Last mutation timestamp.

**Validation Rules**

- `name` must be non-empty after trimming.
- `rate_limit_rps` must be a finite non-negative integer.
- `expires_at` may be omitted or cleared with `null`.
- `metadata` must parse to an object whose values are strings.
- `allowed_models` must preserve exact submitted model identifiers; an empty list means unrestricted.
- Raw API key material is only valid in the create response and must not be persisted in UI state after the create dialog closes.

**State Transitions**

```text
created active -> edited active
created active -> deactivated inactive
inactive -> reactivated active
active or inactive -> deleted
active or inactive -> expired by time
```

## Model Access Rule

Represents the key-scoped allowlist used to restrict which models a key may call.

**Fields**

- `model_id`: Exact identifier submitted in `allowed_models`.
- `label`: Human-readable label composed from available model name, alias, and provider context.
- `is_available`: Whether the model exists in the current model inventory at edit time.

**Validation Rules**

- Empty selected model list means unrestricted access.
- Selected models must submit exact identifiers; display labels cannot replace enforcement identifiers.
- Stale identifiers already present on a key should remain visible as unavailable/stale rather than being silently dropped.

## Key Lifecycle Action

Represents an administrator operation sent from the UI to the gateway.

**Fields**

- `action`: One of `create`, `update`, `deactivate`, `reactivate`, or `delete`.
- `key_id`: Required for all actions except create.
- `payload`: Action-specific request body.
- `status`: `idle`, `submitting`, `succeeded`, `failed`, or `refreshing`.
- `message`: Operator-facing result or error text.

**Validation Rules**

- Delete requires explicit confirmation before submission.
- Deactivate and reactivate are reversible `is_active` updates.
- Failed actions must preserve form state and identify the failed operation.
- Error messages must not include raw key material or sensitive request details.

## Operator Feedback

Represents UI state that communicates lifecycle results and synchronization expectations.

**Fields**

- `severity`: `success`, `info`, `warning`, or `error`.
- `summary`: Short action-specific message.
- `detail`: Optional next step or cache synchronization explanation.
- `refresh_state`: `not_started`, `waiting`, `refreshed`, or `timed_out`.

**Validation Rules**

- Successful mutations trigger key inventory refresh.
- Cache synchronization delays should be explained without implying the backend failed.
- Backend unavailable states should leave the operator with a clear retry path.
