import { test } from "@playwright/test";

// Feature Gap: Settings page shows "Coming soon." placeholder.
// All tests are fixme until the settings surface ships.
// Cross-ref: TESTING.md §8 "Settings".

test.describe("settings page", () => {
  test.fixme("all four settings cards render", async ({ page }) => {
    // Blocked: settings page not yet implemented.
    await page.goto("/settings");
    await page.getByTestId("settings-card-general").isVisible();
    await page.getByTestId("settings-card-security").isVisible();
    await page.getByTestId("settings-card-notifications").isVisible();
    await page.getByTestId("settings-card-advanced").isVisible();
  });
});
