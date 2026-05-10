# Phase 1 Data Model: UI Coming Soon Surfaces

## Request Log Row

Represents one gateway request in the Logs table.

- `id`: stable log row identifier.
- `request_id`: request/call identifier shown to operators.
- `timestamp`: ISO timestamp used for newest-first sorting.
- `method`: HTTP method.
- `path`: request path.
- `status_code`: HTTP status code.
- `latency_ms`: request latency in milliseconds.
- `api_key_id`: key identifier or safe display label; never plaintext key material.
- `model`: requested or resolved model.
- `provider_chain`: ordered `ProviderStage[]` summarizing routing/fallback.

Validation rules:

- `status_code` must be a valid integer HTTP status.
- `latency_ms` must be non-negative.
- `provider_chain` must include at least one stage for fallback-filtered rows.
- Rendered values must pass UI redaction before display.

## Request Inspector Detail

Extends a log row with safe detail tabs.

- `prompt_preview`: redacted, truncated prompt preview.
- `response_preview`: redacted, truncated response preview.
- `headers`: redacted key/value map.
- `trace`: ordered `ProviderStage[]`.

Validation rules:

- Raw prompt/response viewing is out of scope; only preview fields render.
- Authorization, credential, key, token, cookie, and secret-like values must render as `[REDACTED]`.

## Provider Stage

Represents one provider attempt in a request trace or route diagram.

- `provider_id`: provider identifier.
- `provider_name`: display label.
- `outcome`: `success`, `failed`, `skipped`, or `unavailable`.
- `latency_ms`: optional non-negative latency.
- `error_summary`: optional safe error summary.

State transitions:

- A fallback trace may move from failed primary stage to successful secondary stage.
- A route diagram may switch strategy views without mutating runtime routing until saved through a reconciled contract.

## Audit Record

Represents immutable request or admin activity in Audit.

- `id`
- `request_id`
- `actor_id` or actor-like API key identifier.
- `action`
- `api_key_id`
- `model`
- `provider`
- `prompt_tokens`
- `output_tokens`
- `total_tokens`
- `cost_usd`
- `status_code`
- `latency_ms`
- `streaming`
- `error_message`
- `created_at`

Validation rules:

- Records are read-only.
- Filter state supports actor/action/date range and is encoded in URL query parameters.
- Error summaries must be redacted before render.

## Route Policy

Represents one model/provider routing policy.

- `id`
- `model_alias`
- `strategy`: `single`, `fallback`, `weighted`, or `latency`
- `enabled`
- `targets`: `RouteTarget[]`
- `primary_provider_id`
- `fallback_provider_ids`
- `health_summary`

Validation rules:

- `single` requires exactly one target.
- `fallback` requires a primary provider and at least one fallback provider.
- `weighted` requires non-negative weights and a total weight greater than zero.
- Writes remain unavailable in production until the route mutation contract is reconciled.

## Route Target

Represents a provider target within a route policy.

- `provider_id`
- `provider_name`
- `weight`
- `order`
- `health_state`
- `latency_ms`

Validation rules:

- `weight` must be non-negative.
- `order` must be unique within fallback targets when present.

## Guardrail Rule

Represents a gateway safety policy row.

- `id`
- `display_name`
- `enabled`
- `category`
- `description`
- `provider_label`
- `last_updated_at`

Validation rules:

- Toggle is available only for reconciled production contracts or mock-backed E2E contexts.
- Sensitive detector payloads, provider credentials, and raw matched prompts/responses are never displayed.

State transitions:

- `enabled=false` -> saving -> `enabled=true` after successful toggle.
- On failure, the page must preserve or restore state clearly and show the reason.

## Team

Represents a tenant grouping in Teams.

- `id`
- `display_name`
- `member_count`
- `budget_utilization`
- `budget_health`: `ok`, `near_cap`, or `over_cap`

Validation rules:

- `member_count` must be non-negative.
- `budget_utilization` must be non-negative.
- `near_cap` is shown when utilization meets or exceeds the configured alert threshold.

## Budget Policy

Represents a spend control for a team or key.

- `scope`: `team` or `key`
- `scope_id`
- `display_name`
- `limit_usd`
- `usage_usd`
- `duration`: `daily`, `weekly`, `monthly`, or gateway-supported period.
- `alert_threshold`
- `hard_cap_state`: `ok`, `near_cap`, or `over_cap`

Validation rules:

- `limit_usd` and `usage_usd` must be non-negative.
- `alert_threshold` must be greater than 0 and less than or equal to 1.
- Editing policy settings must preserve current usage values.

## Rate Limit Rule

Represents throttling policy across model, key, and team scopes.

- `id`
- `scope`: `model`, `key`, or `team`
- `scope_id`
- `rps`
- `tpm`
- `tripped`

Validation rules:

- `rps` and `tpm`, when present, must be non-negative.
- At least one meaningful limit must be configured for editable forms.
- Token-per-minute controls remain unavailable until production enforcement is confirmed.

## Organization Setting Card

Represents one Settings card.

- `id`: `gateway-profile`, `security`, `notifications`, or `display`.
- `title`
- `description`
- `availability`: `read_only`, `editable`, or `unavailable`.
- `fields`: `SettingField[]`
- `save_capability`
- `last_saved_at`

Validation rules:

- Cards without backing APIs must be disabled/read-only with explanatory copy.
- Secret-like values are redacted or omitted.
- Local-only display preferences may be editable if they persist safely without backend support.

## Setting Field

Represents one value inside a Settings card.

- `id`
- `label`
- `value`
- `input_type`
- `read_only`
- `validation_message`

Validation rules:

- Editable values must preserve user input on validation or save failure.
- Save success must be announced and reflected in visible state.
