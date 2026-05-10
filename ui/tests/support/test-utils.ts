import type { Page } from "@playwright/test";

// Freeze Date.now to a deterministic timestamp (2026-05-09T12:00:00.000Z).
// Call in beforeEach for any test that renders relative-time formatters.
export async function freezeTime(page: Page): Promise<void> {
  await page.addInitScript(
    // This function runs serialized in the browser — plain JS only.
    `Date.now = () => ${1762689600000};`
  );
}

// Disable CSS animations and transitions to stabilize visual snapshots.
// Call before Eyes.check or any screenshot assertion.
export async function disableAnimations(page: Page): Promise<void> {
  await page.addStyleTag({
    content: `*, *::before, *::after { animation: none !important; transition: none !important; }`,
  });
}
