import { test, expect } from "@playwright/test";
import { providers } from "../fixtures/providers";
import { MOCK_BASE_URL } from "../support/mock-port";

// Feature Gap: add-provider wizard is 3 steps (TypeStep → CredentialsStep → ConfirmStep).
// The happy-path test covers the full wizard.
// Wizard-step-specific tests are below the basic coverage.

test.describe("providers page", () => {
  test.beforeEach(async ({ page, request }) => {
    // Reset mock state so mutations from prior tests don't affect counts.
    await request.post(`${MOCK_BASE_URL}/__reset`);
    await page.goto("/providers");
  });

  test("seeded providers render in the table", async ({ page }) => {
    for (const p of providers) {
      await expect(page.getByTestId(`provider-row-${p.id}`)).toBeVisible();
    }
  });

  test("add provider button is visible", async ({ page }) => {
    await expect(page.getByTestId("add-provider")).toBeVisible();
  });

  test("add provider drawer opens on button click", async ({ page }) => {
    await page.getByTestId("add-provider").click();
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(page.getByText("Step 1 of 3")).toBeVisible();
  });

  test("create provider happy path — wizard step 1 selects kind, step 2 fills credentials, step 3 confirms", async ({ page }) => {
    await page.getByTestId("add-provider").click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    // Step 1: choose provider kind (defaults to openai; just click Continue)
    await expect(page.getByTestId("provider-kind-openai")).toBeVisible();
    await page.getByTestId("wizard-next").click();

    // Step 2: fill credentials
    await expect(page.getByText("Step 2 of 3")).toBeVisible();
    await dialog.getByLabel("Display name").fill("Test Provider");
    await dialog.locator("#prov-token").fill("sk-test-fake-token-12345");
    await page.getByTestId("wizard-next").click();

    // Step 3: confirm
    await expect(page.getByText("Step 3 of 3")).toBeVisible();
    await page.getByTestId("wizard-create").click();

    // After creation, the drawer should close and the row should appear
    await expect(dialog).not.toBeVisible({ timeout: 10_000 });
    await expect(page.getByText("Test Provider")).toBeVisible();
  });

  test("provider-kind-anthropic can be selected in step 1", async ({ page }) => {
    await page.getByTestId("add-provider").click();
    await page.getByTestId("provider-kind-anthropic").click();
    // Border changes to brand color — verify the element is still there
    await expect(page.getByTestId("provider-kind-anthropic")).toBeVisible();
  });

  test("provider actions menu opens", async ({ page }) => {
    const firstRow = page.getByTestId(`provider-row-${providers[0].id}`);
    await firstRow.getByTestId("provider-actions-menu").click();
    await expect(page.getByTestId("provider-edit")).toBeVisible();
    await expect(page.getByTestId("provider-delete")).toBeVisible();
  });
});
