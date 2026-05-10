import { test } from "@playwright/test";

// Feature Gap: Audit page shows "Coming soon." placeholder.
// All tests are fixme until the audit surface ships.
// Cross-ref: TESTING.md §7 "Audit".

test.describe("audit page", () => {
  test.fixme("seeded audit entries render in the table", async ({ page }) => {
    // Blocked: audit page not yet implemented.
    await page.goto("/audit");
    // 10 seeded audit entries should render
    await page.locator("[data-testid^='audit-row-']").nth(9).isVisible();
  });

  test.fixme("date filter narrows audit results", async ({ page }) => {
    // Blocked: audit page not yet implemented.
    await page.goto("/audit");
    const fromInput = page.getByTestId("audit-filter-from");
    const toInput = page.getByTestId("audit-filter-to");
    await fromInput.fill("2026-05-09");
    await toInput.fill("2026-05-09");
    await page.getByTestId("audit-filter-apply").click();
    // Results should be filtered
    const rows = page.locator("[data-testid^='audit-row-']");
    const count = await rows.count();
    // At least 1 row, and fewer than all 10
    if (count === 0) throw new Error("Expected filtered results but got none");
  });
});
