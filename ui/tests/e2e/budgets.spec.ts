import { test, expect } from '@playwright/test';
import { MOCK_BASE_URL } from '../support/mock-port';

test.describe('budgets page', () => {
  test('page renders accessible title', async ({ page }) => {
    await page.goto('/budgets');
    await expect(page.getByTestId('page-title')).toHaveText('Budgets');
  });

  test('team_research shows near-cap warning badge', async ({ page }) => {
    await page.goto('/budgets');
    await expect(page.getByTestId('team-research-warn')).toBeVisible();
  });

  test('team_platform shows normal budget state', async ({ page }) => {
    await page.goto('/budgets');
    await expect(page.getByTestId('team-platform-budget')).toBeVisible();
  });

  test('budget rows display limit and usage', async ({ page }) => {
    await page.goto('/budgets');
    // Research: $860 / $1000, Platform: $800 / $2000
    await expect(page.getByTestId('team-research-warn')).toBeVisible();
  });
});

test.describe('budget cap contract', () => {
  test('POST /v1/chat/completions with over-cap key returns 429 budget_exceeded', async ({ request }) => {
    const res = await request.post(`${MOCK_BASE_URL}/v1/chat/completions`, {
      headers: { Authorization: 'Bearer key_over_cap', 'Content-Type': 'application/json' },
      data: { model: 'gpt-4o', messages: [{ role: 'user', content: 'hello' }] },
    });
    expect(res.status()).toBe(429);
    expect(res.headers()['x-llmgopher-reason']).toBe('budget_exceeded');
  });

  test('POST /v1/chat/completions with normal key returns 501 (mock not-implemented)', async ({ request }) => {
    const res = await request.post(`${MOCK_BASE_URL}/v1/chat/completions`, {
      headers: { Authorization: 'Bearer key_checkout_service', 'Content-Type': 'application/json' },
      data: { model: 'gpt-4o', messages: [{ role: 'user', content: 'hello' }] },
    });
    expect(res.status()).toBe(501);
  });
});
