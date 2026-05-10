import { test } from "@playwright/test";

// Feature Gap: Teams page shows "Coming soon." placeholder.
// All tests are fixme until the teams surface ships.
// Cross-ref: specs/23-teams-organizations.

test.describe("teams page", () => {
  test.fixme("teams grid is populated with seeded teams", async ({ page }) => {
    // Blocked: teams page not yet implemented.
    await page.goto("/teams");
    await page.getByTestId("team-row-team_research").isVisible();
    await page.getByTestId("team-row-team_platform").isVisible();
  });

  test.fixme("team_research shows 86% utilization", async ({ page }) => {
    await page.goto("/teams");
    const researchRow = page.getByTestId("team-row-team_research");
    await researchRow.getByText("86%").isVisible();
  });

  test.fixme("team_platform shows 40% utilization", async ({ page }) => {
    await page.goto("/teams");
    const platformRow = page.getByTestId("team-row-team_platform");
    await platformRow.getByText("40%").isVisible();
  });
});
