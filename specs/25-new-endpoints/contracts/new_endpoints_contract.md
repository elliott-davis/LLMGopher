# Contract: Image Generation, Audio & Rerank Endpoints

## Public Surface

This contract captures the externally visible behavior converted from `plans/25-new-endpoints.md`. The exact endpoint, provider adapter, admin route, middleware behavior, storage record, metric, callback, or batch operation is defined by the original plan and refined by `spec.md`.

## Inputs

- Authenticated client, admin, provider, or operator input required to exercise Image Generation, Audio & Rerank Endpoints.
- Runtime configuration and persisted state needed by the feature.
- Provider responses, policy state, cache entries, or external integration events where applicable.

## Outputs

- Public API responses, admin API responses, provider routing outcomes, cache outcomes, observability signals, or audit records produced by the feature.
- Error responses use the established OpenAI-compatible envelope where the public OpenAI-compatible API is involved.
- Admin-only responses avoid leaking credentials, secrets, and provider key material.

## Compatibility Requirements

- Existing gateway clients that do not opt into Image Generation, Audio & Rerank Endpoints continue to work without behavior changes.
- Request IDs, structured logging, audit context, and redaction are preserved.
- Configuration follows CLI flag, environment, config file, and default precedence rules.

## Out of Scope

- Image editing (/v1/images/edits) and variations
- Video generation
- Audio translation (/v1/audio/translations)
- Batch image generation
- Image generation for non-OpenAI providers (Stability AI) in this spec
