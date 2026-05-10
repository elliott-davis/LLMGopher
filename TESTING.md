# Testing Plan — LLMGopher Admin UI

> **Canonical run instructions**: see [`specs/34-ui-e2e-testing-suite/quickstart.md`](specs/34-ui-e2e-testing-suite/quickstart.md).
> This file is the high-level intent document. Refer to `ui/tests/README.md` for directory layout and conventions.

Two layers, both run in CI on every PR:

1. **Visual regression** (Applitools Eyes, Percy, or Chromatic) — catches unintended pixel-level drift across themes, densities, and breakpoints.
2. **Functional E2E** (Playwright) — verifies behavior: drawers open, forms submit, mutations hit the right endpoints, optimistic UI rolls back on failure.

Applitools' Playwright SDK gives you both in one runner — recommended unless you already have Cypress.

---

## Setup

```bash
npm i -D @playwright/test @applitools/eyes-playwright
npx playwright install
npx eyes-setup
```

`playwright.config.ts`: run against a Vite dev server with a deterministic mock backend (MSW or a `--mock` flag on the Go binary that serves the same shapes as `data.js`). Visual tests are non-deterministic if your real backend is in the loop.

```ts
import { defineConfig } from '@playwright/test';
export default defineConfig({
  use: { baseURL: 'http://localhost:5173', viewport: { width: 1440, height: 900 } },
  projects: [
    { name: 'light-comfy',  use: { colorScheme: 'light' } },
    { name: 'dark-comfy',   use: { colorScheme: 'dark'  } },
    { name: 'light-compact' },
  ],
});
```

---

## Visual Regression Suite

One snapshot per screen × theme × density. Use Applitools' **layout regions** for tables and **ignore regions** for live data (timestamps, sparklines, the request flow strip).

### Snapshots to capture

| Screen | States |
|---|---|
| Overview (cards) | empty, populated, error-banner |
| Overview (charts) | populated |
| Overview (tables) | populated |
| Providers | populated, one degraded, one offline |
| Add Provider drawer | step 1, step 2 with creds, step 3 with model picks |
| Routes | each strategy variant (fallback, weighted, latency, single) selected |
| API Keys | populated, key drawer (overview / routes / limits / activity tabs) |
| Logs | mixed status rows, fallback rows visible, empty state |
| Request inspector | trace tab (success), trace tab (fallback), prompt, response, headers |
| Teams | grid populated |
| Budgets | one team near cap, one team over cap |
| Rate limits | one rule "tripped" |
| Guardrails | mix of enabled/disabled |
| Audit | populated |
| Settings | all four cards |

### Suppressing flakiness

```ts
await eyes.check(Target.window().fully()
  .ignoreRegion('.spark')                         // sparklines
  .ignoreRegion('[data-flow-strip]')              // animated pulses
  .ignoreRegion('.id')                            // request ids / timestamps
  .layoutRegion('table.tbl tbody')                // table content shape, not values
);
```

Freeze time in the mock backend and seed a fixed dataset per snapshot. Disable CSS animations during visual runs:

```ts
await page.addStyleTag({ content: `*, *::before, *::after { animation: none !important; transition: none !important; }` });
```

---

## Functional E2E Suite (Playwright)

Group by surface. Selectors should be `data-testid` attributes added to the prototype's interactive nodes (sidebar links, table rows, drawer buttons, form fields).

### 1. Navigation & shell

```ts
test('sidebar navigates between sections', async ({ page }) => {
  await page.goto('/');
  for (const id of ['overview','providers','routes','keys','teams','budgets','rate','logs','guardrails','audit','settings']) {
    await page.getByTestId(`nav-${id}`).click();
    await expect(page.getByTestId('page-title')).toBeVisible();
  }
});

test('⌘K opens command palette', async ({ page }) => {
  await page.keyboard.press('Meta+K');
  await expect(page.getByTestId('command-palette')).toBeVisible();
});
```

### 2. Providers — add flow

```ts
test('add OpenAI provider end-to-end', async ({ page }) => {
  await page.goto('/providers');
  await page.getByTestId('add-provider').click();

  // Step 1
  await page.getByTestId('provider-kind-openai').click();
  await page.getByTestId('wizard-next').click();

  // Step 2
  await page.getByLabel('Display name').fill('OpenAI · prod');
  await page.getByLabel('API key').fill('sk-test-redacted');
  await page.getByTestId('wizard-next').click();
  await expect(page.getByText('connection ok')).toBeVisible({ timeout: 5_000 });

  // Step 3
  await page.getByLabel('gpt-4o').check();
  await page.getByTestId('wizard-create').click();

  await expect(page.getByRole('row', { name: /OpenAI · prod/ })).toBeVisible();
});

test('rejects invalid base URL', async ({ page }) => {
  // ...assert inline error, no POST sent
});
```

### 3. Routes — strategy switching

```ts
test('switching to fallback shows dashed secondary path', async ({ page }) => {
  await page.goto('/routes');
  await page.getByTestId('route-chat-prod').click();
  await expect(page.locator('path[stroke-dasharray="4 4"]')).toHaveCount(1);
});

test('weight slider updates curve thickness', async ({ page }) => {
  // ...
});
```

### 4. API keys — rotate / disable / hard cap

```ts
test('rotate emits one POST and shows new prefix exactly once', async ({ page }) => {
  const requests = [];
  page.on('request', r => r.url().includes('/admin/keys/') && requests.push(r));
  await page.goto('/keys');
  await page.getByRole('row', { name: /checkout-service/ }).click();
  await page.getByTestId('rotate-key').click();
  await page.getByTestId('confirm-rotate').click();
  expect(requests.filter(r => r.method() === 'POST')).toHaveLength(1);
  await expect(page.getByTestId('one-time-key-reveal')).toBeVisible();
});

test('hard-cap toggle shows warn pill once budget > cap', async ({ page }) => {
  // seed key at 99% budget, simulate one more spend tick, expect "near cap" pill flips to danger
});
```

### 5. Logs — filter, click, drawer

```ts
test('filter by 5xx narrows table to error rows only', async ({ page }) => {
  await page.goto('/logs');
  await page.getByTestId('filter-5xx').click();
  for (const row of await page.getByTestId('log-row').all()) {
    await expect(row.getByTestId('status')).toContainText(/5\d\d/);
  }
});

test('inspector trace tab renders waterfall with fallback stage striped', async ({ page }) => {
  await page.goto('/logs');
  await page.getByTestId('log-row-fallback').click();
  const failedBar = page.getByTestId('timeline-stage-primary');
  await expect(failedBar).toHaveAttribute('data-failed', 'true');
});
```

### 6. Budgets — soft + hard cap behavior

```ts
test('alert at 85% threshold shows warning indicator', async ({ page }) => {
  // mock /admin/teams to return Research at 0.86 utilization
  await expect(page.getByTestId('team-research-warn')).toBeVisible();
});

test('hitting hard cap rejects subsequent requests in mock backend', async () => {
  // contract test: POST chat completion with key at cap → 429 with x-llmgopher-reason: budget_exceeded
});
```

### 7. Guardrails — toggle persists

```ts
test('toggling Jailbreak on persists across reload', async ({ page }) => {
  await page.goto('/guardrails');
  await page.getByTestId('toggle-gr_jail').click();
  await page.reload();
  await expect(page.getByTestId('toggle-gr_jail')).toHaveAttribute('aria-checked', 'true');
});
```

### 8. RBAC — role-aware shell

If/when you add the developer-self-serve role:

```ts
test('viewer role hides destructive actions', async ({ page }) => {
  await loginAs(page, { role: 'viewer' });
  await expect(page.getByTestId('rotate-key')).toBeHidden();
  await expect(page.getByTestId('disable-key')).toBeHidden();
});
```

---

## Accessibility

Bake into the same Playwright runner using `@axe-core/playwright`:

```ts
import AxeBuilder from '@axe-core/playwright';
test('overview has no a11y violations', async ({ page }) => {
  await page.goto('/');
  const results = await new AxeBuilder({ page }).analyze();
  expect(results.violations).toEqual([]);
});
```

Critical to verify:
- Sidebar nav has proper `aria-current="page"` on the active item
- Status pills are not the only signal (icons / text labels back them up)
- Drawers trap focus and restore on close
- Tables have proper `<th scope="col">` and sortable headers announce sort state

---

## CI Integration

```yaml
# .github/workflows/ui.yml
- run: npm ci
- run: npm run build
- run: npx playwright test
  env:
    APPLITOOLS_API_KEY: ${{ secrets.APPLITOOLS_API_KEY }}
    APPLITOOLS_BATCH_ID: ${{ github.run_id }}
```

Block PR merge on:
- Any functional test failure
- Any **rejected** visual diff (Applitools dashboard reviewers approve/reject)
- Any new a11y violation

---

## What NOT to test visually

- The live request flow strip (animated, non-deterministic) — assert structure functionally instead (`6 stages rendered`, `pulse spawns at expected interval`)
- Sparklines (random in mocks) — wrap in `.ignoreRegion`
- Timestamps and request IDs — wrap in `.ignoreRegion`
- Spend numbers that depend on real-time aggregation
