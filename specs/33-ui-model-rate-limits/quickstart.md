# Quickstart: UI Model Rate Limit Controls

## Prerequisites

- Gateway and UI dependencies installed.
- Dev stack available through `make dev` when manual smoke testing is needed.
- Existing local admin UI access and gateway admin endpoints configured.

## Implementation Smoke Test

1. Start the local stack:

   ```bash
   make dev
   ```

2. Open the admin UI and navigate to `Models`.

3. Create a model with:

   ```text
   Alias: gpt-4o-limited
   Name: gpt-4o
   Context Window: 128000
   Model Rate Limit: 2
   ```

4. Verify the model inventory shows a model-level limit of `2 requests/sec`.

5. Edit the same model and set the model rate limit to `0`.

6. Verify the model inventory shows an explicit no model-level limit state.

7. Attempt to save a negative model rate limit.

8. Verify the UI rejects the value or displays the gateway rejection while keeping the form state.

## Focused Automated Checks

Run UI tests for the changed model actions and components:

```bash
cd ui
npm test -- --run src/lib/actions.test.ts
```

Run the broader UI verification before implementation is marked complete:

```bash
cd ui
npm test -- --run
npm run lint
npm run build
```

If backend model contract drift is discovered, run focused Go checks:

```bash
go test ./internal/api/... -run Model -v
go test ./internal/storage/... -run Model -v
```

## Expected Results

- Create and update payloads include `rate_limit_rps`.
- Negative values are rejected 100% of the time.
- `0` is clearly communicated as no model-level limit.
- Existing model create, edit, delete, provider assignment, alias, name, and context window workflows still succeed.

## Out of Scope

Token-per-minute limits and per-key-per-model compound limits remain API-only
for this feature. The UI control covers only the existing model-level
`rate_limit_rps` requests-per-second policy.
