export const MOCK_PORT = 8787;
export const MOCK_BASE_URL = `http://127.0.0.1:${MOCK_PORT}`;

// Environment block injected into the Next.js dev-server process when
// LLMGOPHER_UI_BACKEND_MODE=mock. Consumed by playwright.config.ts.
export const mockEnv: Record<string, string> = {
  LLMGOPHER_GATEWAY_BASE: MOCK_BASE_URL,
  LLMGOPHER_UI_ADMIN_API_KEY: "test-token",
  LLMGOPHER_UI_BACKEND_MODE: "mock",
};
