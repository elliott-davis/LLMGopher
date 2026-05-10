import { test } from "@playwright/test";

// Feature Gap: Rate limits page shows "Coming soon." placeholder.
// All tests are fixme until the rate-limits surface from spec 33 ships.
// Cross-ref: specs/33-ui-model-rate-limits.

test.describe("rate limits page", () => {
  test.fixme("rate limit rules are listed", async ({ page }) => {
    // Blocked: rate limits page not yet implemented.
    await page.goto("/rate-limits");
    // 3 seeded rules should render
    await page.locator("[data-testid^='rate-limit-row-']").nth(2).isVisible();
  });

  test.fixme("rule with tripped: true shows tripped pill", async ({ page }) => {
    // Blocked: rate limits page not yet implemented.
    // Cross-ref: rate-limits fixture has exactly 1 tripped rule.
    await page.goto("/rate-limits");
    const trippedPill = page.getByTestId("rate-limit-tripped-pill");
    await trippedPill.isVisible();
  });
});
