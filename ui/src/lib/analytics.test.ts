import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  fetchAuditRecords,
  fetchDailyUsage,
  fetchUsageSummary,
  parseAnalyticsFilters,
} from "@/lib/analytics";

describe("parseAnalyticsFilters", () => {
  it("parses supported filter fields and defaults", () => {
    const filters = parseAnalyticsFilters({
      group_by: "provider",
      from: "2026-04-01T00:00:00Z",
      to: "2026-04-10T00:00:00Z",
      api_key_id: "key-001",
      model: "gpt-4o",
      provider: "openai",
      status: "error",
      limit: "100",
      offset: "50",
    });

    expect(filters.group_by).toBe("provider");
    expect(filters.api_key_id).toBe("key-001");
    expect(filters.status).toBe("error");
    expect(filters.limit).toBe(100);
    expect(filters.offset).toBe(50);
  });

  it("resets offset when non-pagination filters change", () => {
    const filters = parseAnalyticsFilters({
      group_by: "model",
      offset: "75",
      prev_group_by: "model",
      model: "gpt-4.1",
      prev_model: "gpt-4o",
    });

    expect(filters.offset).toBe(0);
  });

  it("throws for invalid date windows", () => {
    expect(() =>
      parseAnalyticsFilters({
        from: "2026-05-01T00:00:00Z",
        to: "2026-04-01T00:00:00Z",
      })
    ).toThrow("from must be before or equal to to");
  });
});

describe("analytics fetch helpers", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    process.env.LLMGOPHER_UI_ADMIN_API_KEY = "admin-token";
  });

  it("sends bearer auth and parses usage responses", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(
        JSON.stringify({
          group_by: "model",
          from: "2026-04-01T00:00:00Z",
          to: "2026-04-30T00:00:00Z",
          data: [
            {
              group: "gpt-4o",
              requests: 10,
              prompt_tokens: 100,
              completion_tokens: 50,
              total_tokens: 150,
              cost_usd: 0.1234,
              errors: 1,
              avg_latency_ms: 420,
            },
          ],
        }),
        { status: 200 }
      )
    );

    const result = await fetchUsageSummary(
      parseAnalyticsFilters({ group_by: "model" })
    );

    expect(result.state).toBe("ready");
    expect(result.data.data).toHaveLength(1);
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining("/v1/admin/usage?"),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer admin-token",
        }),
      })
    );
  });

  it("maps missing admin token to unavailable state", async () => {
    delete process.env.LLMGOPHER_UI_ADMIN_API_KEY;

    const result = await fetchDailyUsage(parseAnalyticsFilters({}));

    expect(result.state).toBe("unavailable");
    expect(result.message).toContain("LLMGOPHER_UI_ADMIN_API_KEY");
  });

  it("maps 400 to invalid-filter", async () => {
    vi.mocked(fetch).mockResolvedValue(new Response("{}", { status: 400 }));

    const result = await fetchAuditRecords(parseAnalyticsFilters({}));

    expect(result.state).toBe("invalid-filter");
  });

  it.each([401, 403, 503])(
    "maps %s responses to unavailable state",
    async (status) => {
      vi.mocked(fetch).mockResolvedValue(new Response("{}", { status }));

      const result = await fetchAuditRecords(parseAnalyticsFilters({}));

      expect(result.state).toBe("unavailable");
    }
  );
});
