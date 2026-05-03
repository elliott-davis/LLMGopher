import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import UsagePage from "@/app/(dashboard)/usage/page";

vi.mock("@/lib/analytics", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/analytics")>()),
  fetchUsageSummary: vi.fn(),
  fetchDailyUsage: vi.fn(),
  fetchAuditRecords: vi.fn(),
}));

import {
  fetchAuditRecords,
  fetchDailyUsage,
  fetchUsageSummary,
} from "@/lib/analytics";

describe("UsagePage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders grouped summary, trend rows, and audit pagination", async () => {
    vi.mocked(fetchUsageSummary).mockResolvedValue({
      state: "ready",
      data: {
        group_by: "model",
        from: "",
        to: "",
        data: [
          {
            group: "gpt-4o",
            requests: 1,
            prompt_tokens: 10,
            completion_tokens: 5,
            total_tokens: 15,
            cost_usd: 0.02,
            errors: 0,
            avg_latency_ms: 100,
          },
        ],
      },
    });
    vi.mocked(fetchDailyUsage).mockResolvedValue({
      state: "ready",
      data: {
        from: "",
        to: "",
        data: [{ date: "2026-04-01", requests: 1, total_tokens: 15, cost_usd: 0.02 }],
      },
    });
    vi.mocked(fetchAuditRecords).mockResolvedValue({
      state: "ready",
      data: {
        data: [
          {
            id: 1,
            request_id: "req-1",
            api_key_id: "key-1",
            model: "gpt-4o",
            provider: "openai",
            prompt_tokens: 10,
            output_tokens: 5,
            total_tokens: 15,
            cost_usd: 0.02,
            status_code: 200,
            latency_ms: 321,
            streaming: false,
            error_message: "",
            created_at: "2026-04-01T00:00:00Z",
          },
        ],
        total: 2,
        limit: 1,
        offset: 0,
      },
    });

    render(
      await UsagePage({
        searchParams: Promise.resolve({
          group_by: "model",
          from: "2026-04-01T00:00:00Z",
        }),
      })
    );

    expect(screen.getByText("Grouped Usage Summary")).toBeInTheDocument();
    expect(screen.getByText("Daily Usage Trend")).toBeInTheDocument();
    expect(screen.getByText("Audit Log Search")).toBeInTheDocument();
    expect(screen.getByText("req-1")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Next" })).toHaveAttribute(
      "href",
      expect.stringContaining("from=2026-04-01T00%3A00%3A00.000Z")
    );
  });

  it("renders empty and unavailable states", async () => {
    vi.mocked(fetchUsageSummary).mockResolvedValue({
      state: "empty",
      data: { group_by: "model", from: "", to: "", data: [] },
    });
    vi.mocked(fetchDailyUsage).mockResolvedValue({
      state: "unavailable",
      data: { from: "", to: "", data: [] },
      message: "Analytics backend is currently unavailable.",
    });
    vi.mocked(fetchAuditRecords).mockResolvedValue({
      state: "empty",
      data: { data: [], total: 0, limit: 25, offset: 0 },
    });

    render(await UsagePage({ searchParams: Promise.resolve({}) }));

    expect(screen.getAllByText("No records matched the current filters.")).toHaveLength(2);
    expect(
      screen.getByText("Analytics backend is currently unavailable.")
    ).toBeInTheDocument();
  });

  it("renders invalid-filter states when filters fail validation", async () => {
    render(
      await UsagePage({
        searchParams: Promise.resolve({
          from: "2026-05-01T00:00:00Z",
          to: "2026-04-01T00:00:00Z",
        }),
      })
    );

    expect(
      screen.getAllByText("from must be before or equal to to")
    ).toHaveLength(3);
    expect(fetchUsageSummary).not.toHaveBeenCalled();
  });
});
