# Quickstart: UI API Key Budget Controls

## Prerequisites

- Docker Compose dev stack dependencies available.
- A gateway admin API key available to the UI server. For local dev, the Compose gateway seeds `sk-test-key-1:key-001`.

## Local Setup

1. Configure the UI service with a server-only admin token for budget endpoints.

   ```bash
   LLMGOPHER_UI_ADMIN_API_KEY=sk-test-key-1
   ```

2. Start the local stack.

   ```bash
   make dev
   ```

3. Open the admin UI at `http://localhost:3000/keys`.

## Manual Validation

1. Confirm the key inventory loads.
2. Open budget controls for an API key with no budget.
3. Verify the UI shows "No budget set" and offers a setup action.
4. Create a valid budget:

   ```text
   limit: 100
   alert threshold: 80
   duration: monthly
   reset time: future datetime
   ```

5. Save and confirm the UI displays limit, spent, remaining, threshold, duration, and reset time.
6. Update the limit or threshold and confirm the spent value is preserved.
7. Reset budget spend, confirm the confirmation dialog appears, and verify spent becomes zero.
8. Remove the budget, confirm the confirmation dialog appears, and verify the UI returns to the no-budget state.
9. Try invalid values and confirm actionable validation feedback:

   ```text
   limit: 0
   threshold: 100
   duration: monthly with no reset time
   ```

10. Remove or invalidate `LLMGOPHER_UI_ADMIN_API_KEY`, restart the UI, and verify budget controls show an authorization/configuration message without exposing the token.
11. Confirm no raw budget API request is made from the browser network panel; budget operations must execute through server actions only.

## Automated Checks

Run UI tests and lint:

```bash
cd ui
npm test
npm run lint
```

If implementation changes gateway budget auth or handlers, run focused Go tests:

```bash
go test ./internal/api/... -run 'Test.*Budget|TestAdminBudget' -v
```

## Expected Results

- Administrators can inspect whether a key has a budget in under 30 seconds.
- Valid budget changes are reflected after the UI refresh flow.
- Reset and remove actions always require explicit confirmation.
- Invalid submissions are rejected with clear feedback.
- Admin credentials remain server-side only.
- Browser-visible traffic never includes bearer auth for budget routes.

## Latest Validation Notes

- 2026-05-02: Automated checks completed with `cd ui && npm test` and `cd ui && npm run lint`.
- 2026-05-02: Manual `make dev` quickstart run not executed in this implementation pass.
