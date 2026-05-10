import { test, expect } from '@playwright/test';

test.describe('settings page', () => {
  test('page renders accessible title', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.getByTestId('page-title')).toHaveText('Settings');
  });

  test('all four settings cards render', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.getByTestId('settings-card-gateway-profile')).toBeVisible();
    await expect(page.getByTestId('settings-card-security')).toBeVisible();
    await expect(page.getByTestId('settings-card-notifications')).toBeVisible();
    await expect(page.getByTestId('settings-card-display')).toBeVisible();
  });

  test('unavailable cards show read-only copy', async ({ page }) => {
    await page.goto('/settings');
    const profileCard = page.getByTestId('settings-card-gateway-profile');
    await expect(profileCard.getByTestId('settings-card-unavailable')).toBeVisible();
  });

  test('display card is editable', async ({ page }) => {
    await page.goto('/settings');
    const displayCard = page.getByTestId('settings-card-display');
    await expect(displayCard.getByTestId('settings-card-save')).toBeVisible();
  });

  test('display card save succeeds', async ({ page }) => {
    await page.goto('/settings');
    const displayCard = page.getByTestId('settings-card-display');
    await displayCard.getByTestId('settings-card-save').click();
    await expect(displayCard.getByTestId('settings-card-success')).toBeVisible();
  });

  test('secret-like field values are redacted', async ({ page }) => {
    await page.goto('/settings');
    const bodyText = await page.textContent('body');
    expect(bodyText).not.toContain('sk-');
  });
});
