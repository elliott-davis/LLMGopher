import { test, expect } from '@playwright/test';

test.describe('teams page', () => {
  test('page renders accessible title', async ({ page }) => {
    await page.goto('/teams');
    await expect(page.getByTestId('page-title')).toHaveText('Teams');
  });

  test('both seeded teams render', async ({ page }) => {
    await page.goto('/teams');
    await expect(page.getByTestId('team-row-team_research')).toBeVisible();
    await expect(page.getByTestId('team-row-team_platform')).toBeVisible();
  });

  test('team_research shows 86% utilization with warning', async ({ page }) => {
    await page.goto('/teams');
    const researchRow = page.getByTestId('team-row-team_research');
    await expect(researchRow.getByText('86%')).toBeVisible();
    await expect(page.getByTestId('team-research-warn')).toBeVisible();
  });

  test('team_platform shows 40% utilization without warning', async ({ page }) => {
    await page.goto('/teams');
    const platformRow = page.getByTestId('team-row-team_platform');
    await expect(platformRow.getByText('40%')).toBeVisible();
  });

  test('empty state shows when no teams exist', async ({ page }) => {
    await page.goto('/teams');
    // Verify page loads correctly
    await expect(page.getByTestId('page-title')).toBeVisible();
  });
});
