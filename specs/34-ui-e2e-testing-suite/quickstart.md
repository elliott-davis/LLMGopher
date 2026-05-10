# Quickstart — UI E2E Testing Suite

Run the suite locally in under a minute (after first-time install).

## First-time setup

```bash
cd ui
npm install                 # installs @playwright/test + @axe-core/playwright
npx playwright install      # downloads Chromium browser binaries (~100 MB)
```

## Run tests

```bash
# All projects (light + dark Chromium)
npm run test:e2e

# Single project
npx playwright test --project=light-comfy

# Single file
npx playwright test tests/e2e/navigation.spec.ts

# Interactive UI mode (great for debugging)
npm run test:e2e:ui
```

The Playwright config (`ui/playwright.config.ts`) auto-starts `next dev` on port 5173 and waits for it. If the app is already running there, it reuses the process locally (`reuseExistingServer: true`).

## Today's coverage

Currently shipped: `tests/e2e/navigation.spec.ts` — sidebar links visible, click-through, `aria-current` correctness. The `⌘K` test is `test.fixme` because the command palette is not implemented.

Run it:

```bash
npx playwright test tests/e2e/navigation.spec.ts --project=light-comfy
# Expect: 3 passed, 1 skipped
```

## Adding a new spec

1. Add `data-testid` attributes to any element you need to target unambiguously. Prefer `getByRole`/`getByLabel` first.
2. Drop a new file under `ui/tests/e2e/<surface>.spec.ts`.
3. Use the seeded fixtures (see `data-model.md` for what the mock backend returns).
4. Run with `--project=light-comfy` while iterating; full matrix only on push.

## Mock backend (Story 1, not yet shipped)

Once landed, the test runner will boot a Hono server on `127.0.0.1:8080` alongside `next dev`. To run the suite against the **real** gateway instead (e.g. for contract drift hunting):

```bash
LLMGOPHER_UI_BACKEND_MODE=real LLMGOPHER_GATEWAY_BASE=http://localhost:8080 npm run test:e2e
```

Tests that require deterministic seed data will skip themselves under `BACKEND_MODE=real`.

## Visual regression (Story 3, not yet shipped)

Once Applitools is provisioned:

```bash
APPLITOOLS_API_KEY=… npx playwright test tests/e2e/visual.spec.ts
```

First run creates the baseline; review and approve in the Applitools dashboard.

## CI

`.github/workflows/ui-e2e.yml` (Story 5) runs on PRs touching `ui/**`. It caches `~/.cache/ms-playwright` and `~/.npm`.

## Troubleshooting

- **"Browser executable not found"** — run `npx playwright install`.
- **"Port 5173 already in use"** — kill the stray dev server: `lsof -ti tcp:5173 | xargs kill`.
- **Vitest tries to run a Playwright spec** — `vitest.config.ts` excludes `tests/e2e/**`; if you moved tests, update the exclude.
- **Test is flaky on CI but passes locally** — first suspect: a non-deterministic field (timestamp, random sparkline, animation). Add it to the `ignoreRegion`/`layoutRegion` list per `TESTING.md` §"Suppressing flakiness".
