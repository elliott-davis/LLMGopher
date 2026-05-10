import { test } from "@playwright/test";

// Feature Gap: Logs page shows "Coming soon." placeholder.
// All tests are fixme until the request-log surface ships.
// Cross-ref: TESTING.md §3 "Logs & Request Inspector".

test.describe("logs page", () => {
  test.fixme("5xx filter narrows table to error rows only", async ({ page }) => {
    await page.goto("/logs");
    await page.getByTestId("filter-status-5xx").click();
    // All visible rows should have error status
    const rows = page.locator("[data-testid^='log-row-']");
    const count = await rows.count();
    for (let i = 0; i < count; i++) {
      await rows.nth(i).getByTestId("status-badge").getAttribute("data-status");
    }
  });

  test.fixme("clicking log_fallback opens inspector with multi-stage timeline", async ({ page }) => {
    await page.goto("/logs");
    await page.getByTestId("log-row-log_fallback").click();
    const inspector = page.getByTestId("request-inspector");
    await inspector.isVisible();
    const primaryStage = inspector.getByTestId("timeline-stage-primary");
    await primaryStage.getAttribute("data-failed");
  });

  test.fixme("request inspector prompt tab shows redacted credential headers", async ({ page }) => {
    await page.goto("/logs");
    await page.getByTestId("log-row-log_fallback").click();
    await page.getByTestId("inspector-tab-headers").click();
    // Authorization header should be redacted
    await page.getByText("[REDACTED]").isVisible();
  });
});
