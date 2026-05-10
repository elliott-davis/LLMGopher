# Data Model: UI Model Rate Limit Controls

## Model Configuration

Represents an existing gateway model row shown and edited in the admin UI.

**Fields**:
- `id`: Stable model identifier.
- `provider_id`: Provider identifier used for routing.
- `name`: Provider-native model name.
- `alias`: Gateway-facing model alias.
- `context_window`: Positive context window size.
- `rate_limit_rps`: Non-negative model-level request limit in requests per second. `0` means no model-level limit.
- `created_at`: Creation timestamp.
- `updated_at`: Last update timestamp.

**Validation rules**:
- `alias`, `name`, and `provider_id` must be non-empty.
- `context_window` must be a positive number.
- `rate_limit_rps` must be an integer number greater than or equal to `0`.
- Existing backend UUID validation for `provider_id` and model `id` remains authoritative.

**Relationships**:
- Belongs to one `Provider`.
- Participates in gateway rate-limit decisions through the state cache after backend synchronization.

## Model Rate Limit

Represents the operator-facing policy value attached to a model.

**Fields**:
- `value`: Non-negative integer from `ModelConfiguration.rate_limit_rps`.
- `unit`: Requests per second.
- `state`: `limited` when `value > 0`; `unrestricted` when `value === 0`.

**Validation rules**:
- Negative values are invalid and must be rejected with clear feedback.
- Empty form input should resolve to the safe default `0` only if the form labels make that default explicit.

**State transitions**:
- `unrestricted` -> `limited`: Administrator sets a positive value and saves.
- `limited` -> `limited`: Administrator changes one positive value to another.
- `limited` -> `unrestricted`: Administrator sets `0` and saves.
- Any state -> unchanged: Save fails; UI keeps entered form state and displays the failure reason.

## Rate Limit Policy Explanation

User-facing copy that distinguishes model-level limits from API key limits.

**Fields**:
- `summary`: Short helper text near create/edit controls.
- `no_limit_label`: Inventory display for `rate_limit_rps === 0`.
- `limited_label`: Inventory display for positive `rate_limit_rps`.

**Validation rules**:
- Must not imply API key limits are changed by this model setting.
- Must use the same unit as the backend field: requests per second.
