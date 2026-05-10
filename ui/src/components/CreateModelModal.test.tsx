import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import CreateModelModal from "@/components/CreateModelModal";
import { Provider } from "@/lib/types";

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
  createModel: vi.fn(),
}));

import { createModel } from "@/lib/actions";

const providers: Provider[] = [
  {
    id: "provider-1",
    name: "OpenAI",
    base_url: "https://api.openai.com/v1",
    auth_type: "bearer",
  },
];

describe("CreateModelModal", () => {
  beforeEach(() => {
    vi.mocked(createModel).mockReset();
  });

  it("renders a default model rate limit input with helper copy", async () => {
    render(<CreateModelModal providers={providers} />);

    await userEvent.click(screen.getByRole("button", { name: "Add Model" }));

    expect(screen.getByLabelText("Model Rate Limit (RPS)")).toHaveValue(0);
    expect(
      screen.getByText(/Use 0 for no model-level limit; API key limits are managed separately./)
    ).toBeInTheDocument();
  });

  it("keeps entered model rate limit visible after a failed save", async () => {
    vi.mocked(createModel).mockRejectedValue(
      new Error("Model rate limit must be a non-negative whole number")
    );
    render(<CreateModelModal providers={providers} />);

    await userEvent.click(screen.getByRole("button", { name: "Add Model" }));
    await userEvent.clear(screen.getByLabelText("Model Rate Limit (RPS)"));
    await userEvent.type(screen.getByLabelText("Model Rate Limit (RPS)"), "-1");
    fireEvent.submit(document.querySelector("form") as HTMLFormElement);

    await waitFor(() => expect(createModel).toHaveBeenCalled());
    expect(
      screen.getByText("Model rate limit must be a non-negative whole number")
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Model Rate Limit (RPS)")).toHaveValue(-1);
  });

  it("surfaces gateway failures without closing the form", async () => {
    vi.mocked(createModel).mockRejectedValue(new Error("gateway rejected model"));
    render(<CreateModelModal providers={providers} />);

    await userEvent.click(screen.getByRole("button", { name: "Add Model" }));
    fireEvent.submit(document.querySelector("form") as HTMLFormElement);

    await waitFor(() => expect(createModel).toHaveBeenCalled());
    expect(screen.getByText("gateway rejected model")).toBeInTheDocument();
    expect(screen.getByRole("dialog", { name: "Create Model" })).toBeInTheDocument();
  });
});
