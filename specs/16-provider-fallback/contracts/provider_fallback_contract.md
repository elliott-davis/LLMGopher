# Contract: Provider Fallback Chains

## Public Surface

This contract captures the externally visible behavior converted from `plans/16-provider-fallback.md`. The exact endpoint, provider adapter, admin route, middleware behavior, storage record, metric, callback, or batch operation is defined by the original plan and refined by `spec.md`.

## Inputs

- Authenticated client, admin, provider, or operator input required to exercise Provider Fallback Chains.
- Runtime configuration and persisted state needed by the feature.
- Provider responses, policy state, cache entries, or external integration events where applicable.

## Outputs

- Public API responses, admin API responses, provider routing outcomes, cache outcomes, observability signals, or audit records produced by the feature.
- Error responses use the established OpenAI-compatible envelope where the public OpenAI-compatible API is involved.
- Admin-only responses avoid leaking credentials, secrets, and provider key material.

## Compatibility Requirements

- Existing gateway clients that do not opt into Provider Fallback Chains continue to work without behavior changes.
- Request IDs, structured logging, audit context, and redaction are preserved.
- Configuration follows CLI flag, environment, config file, and default precedence rules.

## Out of Scope

- Content-policy fallbacks (routing based on the type of error rather than exhausted retries)
- Context-window fallback (spec 18 handles the specific case of context_length_exceeded)
- Global fallback (a catch-all for all models) - use per-model config
