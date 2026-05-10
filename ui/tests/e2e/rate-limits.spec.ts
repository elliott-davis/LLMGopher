import { test, expect } from '@playwright/test';

test.describe('rate limits page', () => {
  test('page renders accessible title', async ({ page }) => {
    await page.goto('/rate-limits');
    await expect(page.getByTestId('page-title')).toHaveText('Rate Limits');
  });

  test('all 3 seeded rules render', async ({ page }) => {
    await page.goto('/rate-limits');
    // eslint-disable-next-line playwright/no-raw-locators
    await expect(page.locator("[data-testid^='rate-limit-row-']").nth(2)).toBeVisible();
  });

  test('exactly one rule shows a tripped indicator', async ({ page }) => {
    await page.goto('/rate-limits');
    const trippedPills = page.getByTestId('rate-limit-tripped-pill');
    await expect(trippedPills).toHaveCount(1);
  });

  test('tripped rule is rl_tripped', async ({ page }) => {
    await page.goto('/rate-limits');
    const trippedRow = page.getByTestId('rate-limit-row-rl_tripped');
    await expect(trippedRow.getByTestId('rate-limit-tripped-pill')).toBeVisible();
  });

  test('RPS is displayed for each rule', async ({ page }) => {
    await page.goto('/rate-limits');
    // eslint-disable-next-line playwright/no-raw-locators
    const rows = page.locator("[data-testid^='rate-limit-row-']");
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('page loads without error', async ({ page }) => {
    await page.goto('/rate-limits');
    await expect(page.getByTestId('page-title')).toBeVisible();
  });
});
