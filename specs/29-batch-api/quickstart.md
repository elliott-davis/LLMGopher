# Quickstart: Batch API Verification

## Prerequisites

- Gateway can run locally with required providers, storage, and configuration for this feature.
- A valid local development API key is available when the feature uses authenticated endpoints.
- Dependencies listed in `plan.md` are implemented or explicitly mocked for verification.

## Automated Checks

```bash
go test ./...
```

## Functional Smoke Test

1. Start the gateway with the normal local development environment.
2. Configure the minimum providers, models, policies, or admin state required by `Batch API`.
3. Exercise the primary success path described in `spec.md` User Story 1.
4. Exercise one negative or unsupported path described in the Edge Cases or Out of Scope sections.
5. Verify responses use the expected public contract and OpenAI-compatible error envelope where applicable.
6. Inspect relevant logs, audit records, metrics, traces, callbacks, cache entries, or budget records for request context and redaction.

## Completion Signal

Mark this feature verified only after automated tests pass and a running-gateway or equivalent integration smoke test has been recorded.
