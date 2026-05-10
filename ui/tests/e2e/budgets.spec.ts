import { test, expect } from "@playwright/test";
import { MOCK_BASE_URL } from "../support/mock-port";

// Feature Gap: Budgets page shows "Coming soon." placeholder.
// The UI tests are fixme; the cap-exceeded contract test hits the mock directly.
// Cross-ref: TESTING.md §5 "Budgets".

test.describe("budgets page", () => {
  test.fixme("team_research shows near-cap warning badge", async ({ page }) => {
    // Blocked: budgets page not yet implemented.
    await page.goto("/budgets");
    const researchCard = page.getByTestId("team-research-warn");
    await expect(researchCard).toBeVisible();
  });

  test.fixme("team_platform shows normal budget state", async ({ page }) => {
    await page.goto("/budgets");
    await page.getByTestId("team-platform-budget").isVisible();
  });
});

test.describe("budget cap contract", () => {
  test("POST /v1/chat/completions with over-cap key returns 429 budget_exceeded", async ({ request }) => {
    const res = await request.post(`${MOCK_BASE_URL}/v1/chat/completions`, {
      headers: {
        Authorization: "Bearer key_over_cap",
        "Content-Type": "application/json",
      },
      data: {
        model: "gpt-4o",
        messages: [{ role: "user", content: "hello" }],
      },
    });
    expect(res.status()).toBe(429);
    expect(res.headers()["x-llmgopher-reason"]).toBe("budget_exceeded");
  });

  test("POST /v1/chat/completions with normal key returns 501 (mock not-implemented)", async ({ request }) => {
    const res = await request.post(`${MOCK_BASE_URL}/v1/chat/completions`, {
      headers: {
        Authorization: "Bearer key_checkout_service",
        "Content-Type": "application/json",
      },
      data: {
        model: "gpt-4o",
        messages: [{ role: "user", content: "hello" }],
      },
    });
    // Mock returns 501 for all non-budget-exceeded chat traffic
    expect(res.status()).toBe(501);
  });
});
