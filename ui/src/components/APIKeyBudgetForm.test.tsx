import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import APIKeyBudgetForm from "@/components/APIKeyBudgetForm";

describe("APIKeyBudgetForm", () => {
  it("submits valid create/update payloads", async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(<APIKeyBudgetForm onSubmit={onSubmit} />);

    await userEvent.type(screen.getByLabelText("Budget limit (USD)"), "100");
    await userEvent.type(screen.getByLabelText("Alert threshold (%)"), "80");
    await userEvent.selectOptions(screen.getByLabelText("Budget duration"), "monthly");
    await userEvent.type(screen.getByLabelText("Reset time"), "2026-06-01T00:00");
    await userEvent.click(screen.getByRole("button", { name: "Save budget" }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    const formData = onSubmit.mock.calls[0][0] as FormData;
    expect(formData.get("budget_usd")).toBe("100");
    expect(formData.get("alert_threshold_pct")).toBe("80");
  });

  it("shows validation errors for invalid values", async () => {
    const onSubmit = vi.fn();
    render(<APIKeyBudgetForm onSubmit={onSubmit} />);

    await userEvent.type(screen.getByLabelText("Budget limit (USD)"), "0");
    await userEvent.type(screen.getByLabelText("Alert threshold (%)"), "100");
    await userEvent.selectOptions(screen.getByLabelText("Budget duration"), "daily");
    await userEvent.click(screen.getByRole("button", { name: "Save budget" }));

    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByText("Enter a budget limit greater than zero.")).toBeInTheDocument();
    expect(screen.getByText("Alert threshold must be 1-99.")).toBeInTheDocument();
    expect(screen.getByText("Reset time is required when duration is set.")).toBeInTheDocument();
  });
});
