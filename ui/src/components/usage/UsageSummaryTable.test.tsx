import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import UsageSummaryTable from "@/components/usage/UsageSummaryTable";

describe("UsageSummaryTable", () => {
  it("renders totals, tokens, errors, latency, and small usd formatting", () => {
    render(
      <UsageSummaryTable
        rows={[
          {
            group: "gpt-4o",
            requests: 12,
            prompt_tokens: 1000,
            completion_tokens: 500,
            total_tokens: 1500,
            cost_usd: 0.009,
            errors: 2,
            avg_latency_ms: 845.5,
          },
        ]}
      />
    );

    expect(screen.getByText("gpt-4o")).toBeInTheDocument();
    expect(screen.getByText("12")).toBeInTheDocument();
    expect(screen.getByText("1,500")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("845.5")).toBeInTheDocument();
    expect(screen.getByText("<$0.01")).toBeInTheDocument();
  });
});
