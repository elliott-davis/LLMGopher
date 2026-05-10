import { test } from "@playwright/test";

// Feature Gap: RBAC / JWT auth is not yet implemented.
// Entire file is fixme until spec 24-rbac-jwt-auth ships.
// Cross-ref: specs/24-rbac-jwt-auth.

test.describe("rbac — viewer role", () => {
  test.fixme("viewer role hides rotate-key action", async ({ page }) => {
    // Blocked: RBAC not yet implemented.
    // When a viewer token is used, the rotate-key menu item should not render.
    await page.goto("/keys");
    const firstRow = page.locator("[data-testid^='key-row-']").first();
    await firstRow.getByTestId("key-actions-menu").click();
    await page.getByTestId("rotate-key").isHidden();
  });

  test.fixme("viewer role hides deactivate-key action", async ({ page }) => {
    // Blocked: RBAC not yet implemented.
    await page.goto("/keys");
    const firstRow = page.locator("[data-testid^='key-row-']").first();
    await firstRow.getByTestId("key-actions-menu").click();
    await page.getByTestId("key-deactivate").isHidden();
  });
});
