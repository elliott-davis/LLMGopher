# Contract: Generic OpenAI-Compatible Provider

## Public Surface

This contract captures the externally visible behavior converted from `plans/11-generic-openai-provider.md`. The exact endpoint, provider adapter, admin route, middleware behavior, storage record, metric, callback, or batch operation is defined by the original plan and refined by `spec.md`.

## Inputs

- Authenticated client, admin, provider, or operator input required to exercise Generic OpenAI-Compatible Provider.
- Runtime configuration and persisted state needed by the feature.
- Provider responses, policy state, cache entries, or external integration events where applicable.

## Outputs

- Public API responses, admin API responses, provider routing outcomes, cache outcomes, observability signals, or audit records produced by the feature.
- Error responses use the established OpenAI-compatible envelope where the public OpenAI-compatible API is involved.
- Admin-only responses avoid leaking credentials, secrets, and provider key material.

## Compatibility Requirements

- Existing gateway clients that do not opt into Generic OpenAI-Compatible Provider continue to work without behavior changes.
- Request IDs, structured logging, audit context, and redaction are preserved.
- Configuration follows CLI flag, environment, config file, and default precedence rules.

## Out of Scope

- Embeddings support via the generic provider (add if needed; OpenAI compat servers often support /embeddings too, but keep this spec focused)
- Streaming format differences (some compat servers have minor SSE deviations - handle in a follow-up)
- Tool/function call translation (generic compat servers use the same format as OpenAI)
