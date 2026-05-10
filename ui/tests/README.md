# UI Tests

## Directory Layout

```
ui/tests/
├── e2e/              # Playwright E2E specs (one file per admin surface)
├── fixtures/         # Deterministic seed data (providers, keys, models, …)
├── mock/
│   ├── handlers/     # Hono route handlers for the in-process mock backend
│   ├── server.ts     # Entry point — mounts all handlers, starts on port 8787
│   ├── state.ts      # In-memory store + reset()
│   └── types.ts      # Type re-exports + mock-only shapes
└── support/
    ├── mock-port.ts  # MOCK_PORT constant + env helper for playwright.config
    └── test-utils.ts # freezeTime(), disableAnimations() helpers
```

## Running the Suite

```bash
# Install Chromium (once)
npm run test:e2e:install

# Run all tests
npm run test:e2e

# Interactive UI mode
npm run test:e2e:ui

# Run a single spec
npx playwright test tests/e2e/providers.spec.ts

# Run all eight replacement surfaces
npm run test:e2e -- tests/e2e/logs.spec.ts tests/e2e/audit.spec.ts tests/e2e/routes.spec.ts tests/e2e/guardrails.spec.ts tests/e2e/teams.spec.ts tests/e2e/budgets.spec.ts tests/e2e/rate-limits.spec.ts tests/e2e/settings.spec.ts
```

## Replacement Surfaces

All eight previously-placeholder pages now have functional implementations and focused E2E coverage.

| Route | Spec file | Key selectors | Notes |
|-------|-----------|---------------|-------|
| `/logs` | `logs.spec.ts` | `log-row-{id}`, `filter-status-{group}`, `request-inspector`, `inspector-tab-{tab}`, `timeline-stage-primary` | Status filter encodes in URL. Inspector fetches detail via `/api/logs/[id]`. |
| `/audit` | `audit.spec.ts` | `audit-row-{id}`, `audit-filter-actor`, `audit-filter-from`, `audit-filter-to`, `audit-filter-apply` | Actor/date filters encode in URL. Newest-first order. Error summaries redacted. |
| `/routes` | `routes.spec.ts` | `route-row-{id}`, `strategy-{name}`, `strategy-view-{name}`, `route-detail-panel`, `route-save-unavailable` | Read-only in production. Mutation returns 501 from mock. |
| `/guardrails` | `guardrails.spec.ts` | `guardrail-row-{id}`, `guardrail-toggle-{id}` | Toggle sends PATCH to mock. Optimistic update with revert on failure. |
| `/teams` | `teams.spec.ts` | `team-row-{id}`, `team-{id}-warn` | `team-research-warn` visible at 86% utilization (threshold 85%). |
| `/budgets` | `budgets.spec.ts` | `team-{scope_id}-budget`, `team-{scope_id}-warn` | Near-cap/over-cap states computed from utilization vs. alert threshold. |
| `/rate-limits` | `rate-limits.spec.ts` | `rate-limit-row-{id}`, `rate-limit-tripped-pill` | Exactly one tripped rule in seed fixtures (`rl_tripped`). |
| `/settings` | `settings.spec.ts` | `settings-card-{id}`, `settings-card-unavailable`, `settings-card-save`, `settings-card-success` | Display card is editable; other cards are unavailable. |

## Conventions (from research.md)

- **All selectors use `data-testid`** — no CSS class or text selectors in `tests/e2e/**`.
  The ESLint rule `no-raw-locators` enforces this.
- **`getByRole` is allowed** for accessible-name assertions (buttons, headings, dialog).
- **Feature gaps use `test.fixme`** — never delete a test; mark it fixme with a comment
  referencing the blocking spec/feature.
- **Mock backend** runs at `http://127.0.0.1:8787`; Next.js is pointed there via
  `LLMGOPHER_GATEWAY_BASE` in `playwright.config.ts`.
- **`POST /__reset`** restores seed state between tests when needed.

## Visual Regression (Applitools)

Visual tests live in `tests/e2e/visual.spec.ts` and require `APPLITOOLS_API_KEY`.

- First run establishes baselines; subsequent runs diff against them.
- Review diffs and approve at https://eyes.applitools.com.
- Flaky regions (timestamps, sparklines) are suppressed with `ignoreRegion`.

## Accessibility (axe-core)

`tests/e2e/a11y.spec.ts` runs `AxeBuilder.analyze()` on every primary route.

- Both `light-comfy` and `dark-comfy` projects must pass.
- Do **not** silently disable axe rules. If a rule legitimately fails, open a design-debt
  issue and reference it in the spec comment.

### Known design-debt rule suppressions

The following rules are temporarily disabled in `a11y.spec.ts`. Each requires a fix in
the UI before the suppression can be removed:

| Rule | Issue | Fix Required |
|------|-------|-------------|
| `color-contrast` | Sidebar + muted-foreground text has insufficient contrast | Design system palette update |
| `empty-table-header` | Action-column `<th>` cells are empty | Add `aria-label` or `scope` to empty headers |
| `heading-order` | Pages skip heading levels (h1 → h3) | Use `h2` in card/section headers |
| `landmark-one-main` | Dashboard pages lack a `<main>` landmark | Wrap dashboard content in `<main>` |
| `region` | Content outside landmark regions | Related to above |
| `scrollable-region-focusable` | Data tables not keyboard-scrollable | Add `tabindex="0"` to overflow containers |
| `select-name` | `<select>` elements lack accessible names | Add `aria-label` or associated `<label>` |

## CI

The `ui-e2e` workflow (`.github/workflows/ui-e2e.yml`) runs on every PR that touches
`ui/**`. Configure branch protection to require this check before merging.

Set `APPLITOOLS_API_KEY` in GitHub Actions secrets to enable visual regression in CI.
