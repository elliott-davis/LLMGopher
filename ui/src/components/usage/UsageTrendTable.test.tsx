import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import UsageTrendTable from "@/components/usage/UsageTrendTable";

describe("UsageTrendTable", () => {
  it("renders date, request count, token count, zero values, and usd", () => {
    render(
      <UsageTrendTable
        rows={[
          { date: "2026-04-01", requests: 4, total_tokens: 1200, cost_usd: 0.4 },
          { date: "2026-04-02", requests: 0, total_tokens: 0, cost_usd: 0 },
        ]}
      />
    );

    expect(screen.getByText("2026-04-01")).toBeInTheDocument();
    expect(screen.getByText("4")).toBeInTheDocument();
    expect(screen.getByText("1,200")).toBeInTheDocument();
    expect(screen.getByText("$0.40")).toBeInTheDocument();
    expect(screen.getAllByText("0")).toHaveLength(2);
    expect(screen.getByText("$0.00")).toBeInTheDocument();
  });
});
