# Contract: RBAC & JWT Authentication

## Public Surface

This contract captures the externally visible behavior converted from `plans/24-rbac-jwt-auth.md`. The exact endpoint, provider adapter, admin route, middleware behavior, storage record, metric, callback, or batch operation is defined by the original plan and refined by `spec.md`.

## Inputs

- Authenticated client, admin, provider, or operator input required to exercise RBAC & JWT Authentication.
- Runtime configuration and persisted state needed by the feature.
- Provider responses, policy state, cache entries, or external integration events where applicable.

## Outputs

- Public API responses, admin API responses, provider routing outcomes, cache outcomes, observability signals, or audit records produced by the feature.
- Error responses use the established OpenAI-compatible envelope where the public OpenAI-compatible API is involved.
- Admin-only responses avoid leaking credentials, secrets, and provider key material.

## Compatibility Requirements

- Existing gateway clients that do not opt into RBAC & JWT Authentication continue to work without behavior changes.
- Request IDs, structured logging, audit context, and redaction are preserved.
- Configuration follows CLI flag, environment, config file, and default precedence rules.

## Out of Scope

- SSO / OIDC integration
- UI-based role management (update the Next.js UI in a separate task)
- Per-endpoint permission tables (role enum is sufficient)
- Token rotation
