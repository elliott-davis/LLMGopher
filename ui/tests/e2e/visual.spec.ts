// Visual regression tests via Applitools Eyes (fixture API).
//
// Requires APPLITOOLS_API_KEY env var (set in .env.local for local dev,
// APPLITOOLS_API_KEY secret in GitHub Actions for CI).
//
// First run establishes baselines; subsequent runs diff against them.
// Review and approve/reject diffs at https://eyes.applitools.com.
//
// Flaky-region suppressions follow TESTING.md §"Suppressing flakiness":
//   .id cells, .spark sparklines, [data-flow-strip], table body timestamps.

import { test } from "@applitools/eyes-playwright/fixture";
import { disableAnimations } from "../support/test-utils";
import { MOCK_BASE_URL } from "../support/mock-port";

// Suppressed selectors: dynamic content that changes every render.
const IGNORE_REGIONS = [
  { selector: ".id" },
  { selector: ".spark" },
  { selector: "[data-flow-strip]" },
];

const LAYOUT_REGIONS = [
  { selector: "table tbody tr td:last-child" },
];

test.describe("visual — overview", () => {
  test("overview / dashboard", async ({ page, eyes }) => {
    await disableAnimations(page);
    await page.goto("/");
    await eyes.check("overview", {
      ignoreRegions: IGNORE_REGIONS,
      layoutRegions: LAYOUT_REGIONS,
    });
  });
});

test.describe("visual — providers", () => {
  test.beforeEach(async ({ request }) => {
    await request.post(`${MOCK_BASE_URL}/__reset`);
  });

  test("providers page — populated", async ({ page, eyes }) => {
    await disableAnimations(page);
    await page.goto("/providers");
    await eyes.check("providers-populated", {
      ignoreRegions: IGNORE_REGIONS,
      layoutRegions: LAYOUT_REGIONS,
    });
  });

  test("add provider drawer open — step 1 (kind selector)", async ({ page, eyes }) => {
    await disableAnimations(page);
    await page.goto("/providers");
    await page.getByTestId("add-provider").click();
    await page.getByRole("dialog").waitFor({ state: "visible" });
    await eyes.check("provider-drawer-step1", {
      ignoreRegions: IGNORE_REGIONS,
    });
  });

  test("add provider drawer open — step 2 (credentials)", async ({ page, eyes }) => {
    await disableAnimations(page);
    await page.goto("/providers");
    await page.getByTestId("add-provider").click();
    await page.getByTestId("wizard-next").click();
    await page.getByText("Step 2 of 3").waitFor({ state: "visible" });
    await eyes.check("provider-drawer-step2", {
      ignoreRegions: IGNORE_REGIONS,
    });
  });
});

test.describe("visual — keys", () => {
  test("keys page — populated", async ({ page, eyes }) => {
    await disableAnimations(page);
    await page.goto("/keys");
    await eyes.check("keys-populated", {
      ignoreRegions: IGNORE_REGIONS,
      layoutRegions: LAYOUT_REGIONS,
    });
  });
});

test.describe("visual — models", () => {
  test("models page — populated", async ({ page, eyes }) => {
    await disableAnimations(page);
    await page.goto("/models");
    await eyes.check("models-populated", {
      ignoreRegions: IGNORE_REGIONS,
      layoutRegions: LAYOUT_REGIONS,
    });
  });
});

// Remaining snapshot targets (add as pages ship):
// - Routes (each strategy variant)  → test.fixme until routes page ships
// - Logs (mixed / fallback / empty) → test.fixme until logs page ships
// - Teams, Budgets, Rate limits, Guardrails, Audit, Settings → test.fixme
