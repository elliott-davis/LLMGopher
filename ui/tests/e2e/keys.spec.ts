import { test, expect } from "@playwright/test";
import { keys } from "../fixtures/keys";
import { MOCK_BASE_URL } from "../support/mock-port";

// Feature Gap: Key rotation and one-time-reveal are not yet implemented in the UI.
// Those tests are marked test.fixme.

test.describe("keys page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/keys");
  });

  test("seeded keys render in the table", async ({ page }) => {
    for (const k of keys) {
      await expect(page.getByTestId(`key-row-${k.id}`)).toBeVisible();
    }
  });

  test("key with active status shows 'Active'", async ({ page }) => {
    const checkoutRow = page.getByTestId("key-row-key_checkout_service");
    await expect(checkoutRow.getByText("Active")).toBeVisible();
  });

  test("key actions menu opens", async ({ page }) => {
    const firstRow = page.getByTestId(`key-row-${keys[0].id}`);
    await firstRow.getByTestId("key-actions-menu").click();
    await expect(page.getByTestId("key-deactivate")).toBeVisible();
    await expect(page.getByTestId("key-delete")).toBeVisible();
  });

  test("budget cap contract — key_over_cap returns 429 with budget_exceeded reason", async ({ request }) => {
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

  test.fixme("rotate key — emits exactly one POST to /admin/keys/{id}/rotate", async ({ page }) => {
    // Blocked: key rotation not yet implemented in APIKeyRowActions.
    // Cross-ref: Feature Gap "Rotate Key" in TESTING.md.
    const firstRow = page.getByTestId(`key-row-${keys[0].id}`);
    await firstRow.getByTestId("key-actions-menu").click();
    await page.getByTestId("rotate-key").click();
    await page.getByTestId("confirm-rotate").click();
  });

  test.fixme("one-time key reveal shown exactly once after rotate", async ({ page }) => {
    // Blocked: one-time-reveal UI not yet implemented.
    // Cross-ref: Feature Gap "Key Rotation + One-Time Reveal" in TESTING.md.
    const firstRow = page.getByTestId(`key-row-${keys[0].id}`);
    await firstRow.getByTestId("key-actions-menu").click();
    await page.getByTestId("rotate-key").click();
    await page.getByTestId("confirm-rotate").click();
    const reveal = page.getByTestId("one-time-key-reveal");
    await expect(reveal).toBeVisible();
    // Second render shouldn't show it again
    await page.reload();
    await expect(page.getByTestId("one-time-key-reveal")).not.toBeVisible();
  });

  test.fixme("hard-cap toggle flips warn-pill to danger when budget crosses cap", async ({ page }) => {
    // Blocked: hard-cap toggle UI not yet implemented.
    // Cross-ref: Feature Gap "Hard-cap toggle" in TESTING.md.
    const nearCapRow = page.getByTestId("key-row-key_near_cap");
    await nearCapRow.getByTestId("key-hard-cap-toggle").click();
    // After toggling, the budget pill should show danger styling
    await expect(nearCapRow.getByTestId("budget-pill")).toHaveAttribute("data-variant", "danger");
  });
});
