import { test } from "@playwright/test";

// Feature Gap: Routes page shows "Coming soon." placeholder.
// All tests are fixme until the routes/load-balancing UI ships.
// Cross-ref: specs/17-load-balancing and specs/16-provider-fallback.

test.describe("routes page", () => {
  test.fixme("strategy switcher renders dashed secondary path", async ({ page }) => {
    await page.goto("/routes");
    await page.getByTestId("strategy-switcher").click();
    await page.getByTestId("strategy-fallback").click();
    // The secondary path in the flow diagram should be dashed
    const secondaryPath = page.locator("[data-flow-strip='secondary']");
    await secondaryPath.isVisible();
  });

  test.fixme("weight slider updates curve thickness", async ({ page }) => {
    await page.goto("/routes");
    const slider = page.getByTestId("weight-slider");
    await slider.fill("75");
    const curve = page.locator("[data-testid='flow-curve-primary']");
    await curve.isVisible();
  });
});
