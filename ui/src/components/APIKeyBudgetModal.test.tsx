import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import APIKeyBudgetModal from "@/components/APIKeyBudgetModal";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    refresh: vi.fn(),
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/lib/actions", () => ({
  deleteAPIKeyBudget: vi.fn(),
  resetAPIKeyBudget: vi.fn(),
  upsertAPIKeyBudget: vi.fn(),
}));

import {
  deleteAPIKeyBudget,
  resetAPIKeyBudget,
  upsertAPIKeyBudget,
} from "@/lib/actions";

describe("APIKeyBudgetModal", () => {
  beforeEach(() => {
    vi.stubGlobal("confirm", vi.fn(() => true));
  });

  it("submits budget updates and keeps reset/remove available", async () => {
    vi.mocked(upsertAPIKeyBudget).mockResolvedValue({
      api_key_id: "key-1",
      budget_usd: 100,
      spent_usd: 10,
      remaining_usd: 90,
      alert_threshold_pct: 80,
    });

    render(
      <APIKeyBudgetModal
        apiKeyID="key-1"
        apiKeyName="Production"
        initialState={{ status: "unbudgeted" }}
        open
        onOpenChange={() => undefined}
      />
    );

    await userEvent.type(screen.getByLabelText("Budget limit (USD)"), "100");
    await userEvent.click(screen.getByRole("button", { name: "Save budget" }));

    await waitFor(() => expect(upsertAPIKeyBudget).toHaveBeenCalled());
    expect(screen.getByRole("button", { name: "Reset spend" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "Remove budget" })).toBeEnabled();
  });

  it("requires confirmation for reset and remove actions", async () => {
    vi.mocked(resetAPIKeyBudget).mockResolvedValue({
      api_key_id: "key-1",
      budget_usd: 100,
      spent_usd: 0,
      remaining_usd: 100,
    });
    vi.mocked(deleteAPIKeyBudget).mockResolvedValue(undefined);

    render(
      <APIKeyBudgetModal
        apiKeyID="key-1"
        apiKeyName="Production"
        initialState={{
          status: "configured",
          budget: {
            api_key_id: "key-1",
            budget_usd: 100,
            spent_usd: 25,
            remaining_usd: 75,
          },
        }}
        open
        onOpenChange={() => undefined}
      />
    );

    await userEvent.click(screen.getByRole("button", { name: "Reset spend" }));
    await waitFor(() => expect(resetAPIKeyBudget).toHaveBeenCalled());

    await userEvent.click(screen.getByRole("button", { name: "Remove budget" }));
    await waitFor(() => expect(deleteAPIKeyBudget).toHaveBeenCalled());
    expect(screen.getByText("No budget set")).toBeInTheDocument();
  });
});
