import { test, expect } from '@playwright/test';

test.describe('guardrails page', () => {
  test('page renders accessible title', async ({ page }) => {
    await page.goto('/guardrails');
    await expect(page.getByTestId('page-title')).toHaveText('Guardrails');
  });

  test('seeded guardrail rows are listed', async ({ page }) => {
    await page.goto('/guardrails');
    await expect(page.getByTestId('guardrail-row-gr_jail')).toBeVisible();
    await expect(page.getByTestId('guardrail-row-gr_pii')).toBeVisible();
    await expect(page.getByTestId('guardrail-row-gr_secrets')).toBeVisible();
  });

  test('gr_pii starts as enabled', async ({ page }) => {
    await page.goto('/guardrails');
    await expect(page.getByTestId('guardrail-toggle-gr_pii')).toHaveAttribute('aria-checked', 'true');
  });

  test('gr_jail starts as disabled', async ({ page }) => {
    await page.goto('/guardrails');
    await expect(page.getByTestId('guardrail-toggle-gr_jail')).toHaveAttribute('aria-checked', 'false');
  });

  test('toggling gr_jail enables it', async ({ page }) => {
    await page.goto('/guardrails');
    await page.getByTestId('guardrail-toggle-gr_jail').click();
    await expect(page.getByTestId('guardrail-toggle-gr_jail')).toHaveAttribute('aria-checked', 'true');
  });

  test('page loads without error', async ({ page }) => {
    await page.goto('/guardrails');
    await expect(page.getByTestId('page-title')).toBeVisible();
  });
});
