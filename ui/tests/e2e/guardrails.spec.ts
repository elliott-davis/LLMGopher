import { test } from "@playwright/test";

// Feature Gap: Guardrails page shows "Coming soon." placeholder.
// All tests are fixme until the guardrails surface ships.
// Cross-ref: specs/26-guardrail-integrations.

test.describe("guardrails page", () => {
  test.fixme("seeded guardrails are listed", async ({ page }) => {
    // Blocked: guardrails page not yet implemented.
    await page.goto("/guardrails");
    await page.getByTestId("guardrail-row-gr_jail").isVisible();
    await page.getByTestId("guardrail-row-gr_pii").isVisible();
    await page.getByTestId("guardrail-row-gr_secrets").isVisible();
  });

  test.fixme("toggling gr_jail to enabled persists across reload", async ({ page }) => {
    // Blocked: guardrails page not yet implemented.
    // gr_jail starts as enabled=false in fixtures; toggling should persist.
    await page.goto("/guardrails");
    const jailToggle = page.getByTestId("guardrail-toggle-gr_jail");
    await jailToggle.click();
    await page.reload();
    await jailToggle.getAttribute("aria-checked");
    // expect aria-checked="true" after reload
  });
});
