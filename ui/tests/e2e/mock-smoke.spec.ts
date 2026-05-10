import { test, expect } from "@playwright/test";
import { MOCK_BASE_URL } from "../support/mock-port";
import { providers } from "../fixtures/providers";
import { keys } from "../fixtures/keys";
import { teams } from "../fixtures/teams";
import { rateLimits } from "../fixtures/rate-limits";
import { guardrails } from "../fixtures/guardrails";
import { logs } from "../fixtures/logs";
import { auditEntries } from "../fixtures/audit";

// US1 independent test gate: the mock backend returns seeded data for every admin surface.

test.describe("mock backend smoke", () => {
  // Reset state before each test so mutations from other specs don't bleed in.
  test.beforeEach(async ({ request }) => {
    await request.post(`${MOCK_BASE_URL}/__reset`);
  });

  // Providers and Keys return raw arrays (matching real gateway).
  test("GET /v1/admin/providers returns seeded providers", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/providers`);
    expect(res.status()).toBe(200);
    const body = await res.json() as unknown[];
    expect(body).toHaveLength(providers.length);
  });

  test("GET /v1/admin/keys returns seeded keys", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/keys`);
    expect(res.status()).toBe(200);
    const body = await res.json() as unknown[];
    expect(body).toHaveLength(keys.length);
  });

  // Teams, rate-limits, guardrails return { data: [...] }.
  test("GET /v1/admin/teams returns seeded teams", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/teams`);
    expect(res.status()).toBe(200);
    const body = await res.json() as { data: unknown[] };
    expect(body.data).toHaveLength(teams.length);
  });

  test("GET /v1/admin/rate-limits returns seeded rules", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/rate-limits`);
    expect(res.status()).toBe(200);
    const body = await res.json() as { data: unknown[] };
    expect(body.data).toHaveLength(rateLimits.length);
  });

  test("GET /v1/admin/guardrails returns seeded guardrails", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/guardrails`);
    expect(res.status()).toBe(200);
    const body = await res.json() as { data: unknown[] };
    expect(body.data).toHaveLength(guardrails.length);
  });

  test("GET /v1/admin/logs returns seeded log rows", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/logs`);
    expect(res.status()).toBe(200);
    const body = await res.json() as { data: unknown[] };
    expect(body.data).toHaveLength(logs.length);
  });

  // Audit returns AuditResponse { data, total, limit, offset }.
  test("GET /v1/admin/audit returns seeded audit entries", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/audit`);
    expect(res.status()).toBe(200);
    const body = await res.json() as { data: unknown[]; total: number };
    expect(body.data).toHaveLength(auditEntries.length);
    expect(body.total).toBe(auditEntries.length);
  });

  test("x-mock-error header returns the requested error envelope", async ({ request }) => {
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/providers`, {
      headers: { "x-mock-error": "rate_limit_error" },
    });
    expect(res.status()).toBe(429);
    const body = await res.json() as { error: { type: string } };
    expect(body.error.type).toBe("rate_limit_error");
  });

  test("POST /__reset restores seed state", async ({ request }) => {
    // Mutate something
    await request.post(`${MOCK_BASE_URL}/v1/admin/providers`, {
      data: { kind: "generic", name: "Temp provider" },
    });
    // Reset
    await request.post(`${MOCK_BASE_URL}/__reset`);
    const res = await request.get(`${MOCK_BASE_URL}/v1/admin/providers`);
    const body = await res.json() as unknown[];
    expect(body).toHaveLength(providers.length);
  });

  test("budget cap returns 429 with x-llmgopher-reason header", async ({ request }) => {
    const res = await request.post(`${MOCK_BASE_URL}/v1/chat/completions`, {
      headers: { Authorization: "Bearer key_over_cap", "Content-Type": "application/json" },
      data: { model: "gpt-4o", messages: [{ role: "user", content: "hello" }] },
    });
    expect(res.status()).toBe(429);
    expect(res.headers()["x-llmgopher-reason"]).toBe("budget_exceeded");
  });
});
