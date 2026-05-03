import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import APIKeyBudgetStatus from "@/components/APIKeyBudgetStatus";

describe("APIKeyBudgetStatus", () => {
  it("renders configured and threshold states", () => {
    const { rerender } = render(
      <APIKeyBudgetStatus
        state={{
          status: "configured",
          budget: {
            api_key_id: "key-1",
            budget_usd: 100,
            spent_usd: 30,
            remaining_usd: 70,
            alert_threshold_pct: 80,
          },
        }}
      />
    );
    expect(screen.getByText("Within budget")).toBeInTheDocument();

    rerender(
      <APIKeyBudgetStatus
        state={{
          status: "configured",
          budget: {
            api_key_id: "key-1",
            budget_usd: 100,
            spent_usd: 80,
            remaining_usd: 20,
            alert_threshold_pct: 80,
          },
        }}
      />
    );
    expect(screen.getByText("Near threshold")).toBeInTheDocument();
  });

  it("renders exhausted, unbudgeted, and unavailable states", () => {
    const { rerender } = render(
      <APIKeyBudgetStatus
        state={{
          status: "configured",
          budget: {
            api_key_id: "key-1",
            budget_usd: 100,
            spent_usd: 100,
            remaining_usd: 0,
          },
        }}
      />
    );
    expect(screen.getByText("Exhausted")).toBeInTheDocument();

    rerender(<APIKeyBudgetStatus state={{ status: "unbudgeted" }} />);
    expect(screen.getByText("No budget set")).toBeInTheDocument();

    rerender(
      <APIKeyBudgetStatus state={{ status: "unavailable", message: "Budget unavailable" }} />
    );
    expect(screen.getByText("Budget unavailable")).toBeInTheDocument();
  });
});
