# LLMGopher Admin UI

## Local Development

Run the UI directly:

```bash
cd ui
npm install
npm run dev
```

Run with the full local stack:

```bash
LLMGOPHER_UI_ADMIN_API_KEY=sk-test-key-1 make dev
```

Open `http://localhost:3000/keys`.

## Budget Controls

Budget server actions call protected gateway routes and require a server-only admin
token:

```bash
LLMGOPHER_UI_ADMIN_API_KEY=sk-test-key-1
```

- Never pass this token to client-side code or browser storage.
- If missing, budget status shows as unavailable and mutation actions return a setup
  message.
- Budget validation mirrors gateway rules:
  - `budget_usd` > 0
  - `alert_threshold_pct` optional, 1-99
  - `budget_duration` optional (`daily|weekly|monthly`)
  - `budget_reset_at` required when duration is set

## Validation Commands

```bash
cd ui
npm test
npm run lint
```

For budget manual checks, follow `specs/31-ui-key-budgets/quickstart.md`.
