# Contract: Prometheus Metrics Endpoint

## Public Surface

This contract captures the externally visible behavior converted from `plans/20-prometheus-metrics.md`. The exact endpoint, provider adapter, admin route, middleware behavior, storage record, metric, callback, or batch operation is defined by the original plan and refined by `spec.md`.

## Inputs

- Authenticated client, admin, provider, or operator input required to exercise Prometheus Metrics Endpoint.
- Runtime configuration and persisted state needed by the feature.
- Provider responses, policy state, cache entries, or external integration events where applicable.

## Outputs

- Public API responses, admin API responses, provider routing outcomes, cache outcomes, observability signals, or audit records produced by the feature.
- Error responses use the established OpenAI-compatible envelope where the public OpenAI-compatible API is involved.
- Admin-only responses avoid leaking credentials, secrets, and provider key material.

## Compatibility Requirements

- Existing gateway clients that do not opt into Prometheus Metrics Endpoint continue to work without behavior changes.
- Request IDs, structured logging, audit context, and redaction are preserved.
- Configuration follows CLI flag, environment, config file, and default precedence rules.

## Out of Scope

- OpenTelemetry metrics export (spec 21)
- Per-user or per-team metrics (too high cardinality for Prometheus labels)
- Dashboard provisioning (a companion Grafana dashboard JSON can be added as a separate file)
