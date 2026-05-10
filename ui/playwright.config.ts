import { defineConfig, devices } from "@playwright/test";
import { MOCK_PORT, mockEnv } from "./tests/support/mock-port";

const APP_PORT = Number(process.env.PLAYWRIGHT_PORT ?? 5173);

export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 30_000,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL: `http://localhost:${APP_PORT}`,
    viewport: { width: 1440, height: 900 },
    trace: "on-first-retry",
    eyesConfig: {
      appName: "LLMGopher Admin UI",
      batchName: process.env.APPLITOOLS_BATCH_NAME ?? "LLMGopher E2E",
      // Batch ID is set per CI run so all visual tests in a run share one batch.
      ...(process.env.APPLITOOLS_BATCH_ID
        ? { batch: { id: process.env.APPLITOOLS_BATCH_ID } }
        : {}),
    },
  },
  projects: [
    {
      name: "light-comfy",
      use: { ...devices["Desktop Chrome"], colorScheme: "light" },
    },
    {
      name: "dark-comfy",
      use: { ...devices["Desktop Chrome"], colorScheme: "dark" },
    },
  ],
  webServer: [
    {
      // Boot mock gateway first so Next.js can reach it during server-component rendering.
      command: `npx tsx tests/mock/server.ts`,
      url: `http://127.0.0.1:${MOCK_PORT}/healthz`,
      reuseExistingServer: !process.env.CI,
      timeout: 30_000,
    },
    {
      command: `npm run dev -- --port ${APP_PORT}`,
      url: `http://localhost:${APP_PORT}`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      env: {
        ...mockEnv,
        // Pass APPLITOOLS_API_KEY through if set; ignored when running functional tests.
        ...(process.env.APPLITOOLS_API_KEY
          ? { APPLITOOLS_API_KEY: process.env.APPLITOOLS_API_KEY }
          : {}),
      },
    },
  ],
});
