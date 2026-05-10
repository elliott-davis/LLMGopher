import { test, expect } from "@playwright/test";

const NAV_IDS = [
  "overview",
  "logs",
  "audit",
  "providers",
  "routes",
  "guardrails",
  "keys",
  "teams",
  "budgets",
  "rate",
  "settings",
] as const;

test.describe("shell navigation", () => {
  test("sidebar renders every section link", async ({ page }) => {
    await page.goto("/");
    for (const id of NAV_IDS) {
      await expect(page.getByTestId(`nav-${id}`)).toBeVisible();
    }
  });

  test("clicking a nav item navigates and updates breadcrumb", async ({ page }) => {
    await page.goto("/");
    for (const id of NAV_IDS) {
      await page.getByTestId(`nav-${id}`).click();
      await expect(page.getByTestId("page-title")).toBeVisible();
    }
  });

  test("active nav item exposes aria-current=page", async ({ page }) => {
    await page.goto("/providers");
    await expect(page.getByTestId("nav-providers")).toHaveAttribute(
      "aria-current",
      "page",
    );
  });
});

test.describe("command palette", () => {
  test.fixme(
    "⌘K opens command palette",
    // Blocked: command palette is not implemented yet.
    // Tracked in specs/34-ui-e2e-testing-suite as feature gap "command-palette".
    async ({ page }) => {
      await page.goto("/");
      await page.keyboard.press("Meta+K");
      await expect(page.getByTestId("command-palette")).toBeVisible();
    },
  );
});
