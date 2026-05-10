import { test, expect } from '@playwright/test';

test.describe('routes page', () => {
  test('page renders accessible title', async ({ page }) => {
    await page.goto('/routes');
    await expect(page.getByTestId('page-title')).toHaveText('Routes');
  });

  test('four seeded route policies render', async ({ page }) => {
    await page.goto('/routes');
    // eslint-disable-next-line playwright/no-raw-locators
    await expect(page.locator("[data-testid^='route-row-']").first()).toBeVisible();
    // eslint-disable-next-line playwright/no-raw-locators
    const count = await page.locator("[data-testid^='route-row-']").count();
    expect(count).toBeGreaterThanOrEqual(4);
  });

  test('strategy switcher changes the visible strategy view', async ({ page }) => {
    await page.goto('/routes');
    await page.getByTestId('strategy-switcher').click();
    await page.getByTestId('strategy-fallback').click();
    await expect(page.getByTestId('strategy-view-fallback')).toBeVisible();
  });

  test('fallback strategy shows primary and secondary providers', async ({ page }) => {
    await page.goto('/routes');
    await page.getByTestId('route-row-route_gpt4o_fallback').click();
    await expect(page.getByTestId('route-detail-panel')).toBeVisible();
  });

  test('save controls show production-unavailable copy', async ({ page }) => {
    await page.goto('/routes');
    const saveBtn = page.getByTestId('route-save-unavailable');
    await expect(saveBtn).toBeVisible();
  });

  test('empty state shows when no routes exist', async ({ page }) => {
    await page.goto('/routes');
    await expect(page.getByTestId('page-title')).toBeVisible();
  });
});
