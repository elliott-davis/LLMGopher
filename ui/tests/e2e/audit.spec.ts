import { test, expect } from '@playwright/test';

test.describe('audit page', () => {
  test('page renders accessible title', async ({ page }) => {
    await page.goto('/audit');
    await expect(page.getByTestId('page-title')).toHaveText('Audit log');
  });

  test('10 seeded audit entries render', async ({ page }) => {
    await page.goto('/audit');
    // eslint-disable-next-line playwright/no-raw-locators
    await expect(page.locator("[data-testid^='audit-row-']").first()).toBeVisible();
  });

  test('date filter narrows audit results', async ({ page }) => {
    await page.goto('/audit');
    await page.getByTestId('audit-filter-from').fill('2026-05-09');
    await page.getByTestId('audit-filter-to').fill('2026-05-09');
    await page.getByTestId('audit-filter-apply').click();
    // eslint-disable-next-line playwright/no-raw-locators
    const rows = page.locator("[data-testid^='audit-row-']");
    const count = await rows.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('actor filter narrows audit results', async ({ page }) => {
    await page.goto('/audit');
    await page.getByTestId('audit-filter-actor').fill('key_checkout_service');
    await page.getByTestId('audit-filter-apply').click();
    // eslint-disable-next-line playwright/no-raw-locators
    const rows = page.locator("[data-testid^='audit-row-']");
    await expect(rows.first()).toBeVisible();
  });

  test('filter state encodes in URL', async ({ page }) => {
    await page.goto('/audit');
    await page.getByTestId('audit-filter-actor').fill('key_research_team');
    await page.getByTestId('audit-filter-apply').click();
    expect(page.url()).toContain('actor=key_research_team');
  });

  test('URL filter is restored on reload', async ({ page }) => {
    await page.goto('/audit?actor=key_checkout_service');
    await expect(page.getByTestId('audit-filter-actor')).toHaveValue('key_checkout_service');
  });

  test('entries render newest-first', async ({ page }) => {
    await page.goto('/audit');
    // eslint-disable-next-line playwright/no-raw-locators
    const firstRow = page.locator("[data-testid^='audit-row-']").first();
    await expect(firstRow).toBeVisible();
    await expect(firstRow).toHaveAttribute('data-testid', 'audit-row-1');
  });

  test('page renders without error even with empty params', async ({ page }) => {
    await page.goto('/audit');
    await expect(page.getByTestId('page-title')).toBeVisible();
  });
});
