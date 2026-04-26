# Contract: Additional OpenAI-Compatible Providers

## Public Surface

This contract captures the externally visible behavior converted from `plans/14-additional-providers.md`. The exact endpoint, provider adapter, admin route, middleware behavior, storage record, metric, callback, or batch operation is defined by the original plan and refined by `spec.md`.

## Inputs

- Authenticated client, admin, provider, or operator input required to exercise Additional OpenAI-Compatible Providers.
- Runtime configuration and persisted state needed by the feature.
- Provider responses, policy state, cache entries, or external integration events where applicable.

## Outputs

- Public API responses, admin API responses, provider routing outcomes, cache outcomes, observability signals, or audit records produced by the feature.
- Error responses use the established OpenAI-compatible envelope where the public OpenAI-compatible API is involved.
- Admin-only responses avoid leaking credentials, secrets, and provider key material.

## Compatibility Requirements

- Existing gateway clients that do not opt into Additional OpenAI-Compatible Providers continue to work without behavior changes.
- Request IDs, structured logging, audit context, and redaction are preserved.
- Configuration follows CLI flag, environment, config file, and default precedence rules.

## Out of Scope

- Custom request format per provider beyond what the generic compat provider handles
- Provider-specific embeddings (Mistral has embeddings - add in a follow-up)
- OpenRouter-specific multi-provider routing (treat as a single provider)
