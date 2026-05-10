import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import EditModelModal from "@/components/EditModelModal";
import { Model, Provider } from "@/lib/types";

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
  updateModel: vi.fn(),
}));

import { updateModel } from "@/lib/actions";

const model: Model = {
  id: "model-1",
  alias: "gpt-4o",
  name: "gpt-4o-2024-11-20",
  provider_id: "provider-1",
  context_window: 128000,
  rate_limit_rps: 12,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const providers: Provider[] = [
  {
    id: "provider-1",
    name: "OpenAI",
    base_url: "https://api.openai.com/v1",
    auth_type: "bearer",
  },
];

describe("EditModelModal", () => {
  beforeEach(() => {
    vi.mocked(updateModel).mockReset();
  });

  it("renders the existing model rate limit with helper copy", () => {
    render(
      <EditModelModal
        model={model}
        providers={providers}
        open
        onOpenChange={() => undefined}
      />
    );

    expect(screen.getByLabelText("Model Rate Limit (RPS)")).toHaveValue(12);
    expect(
      screen.getByText(/Use 0 for no model-level limit; API key limits are managed separately./)
    ).toBeInTheDocument();
  });

  it("submits changed rate limit values", async () => {
    vi.mocked(updateModel).mockResolvedValue(undefined);
    render(
      <EditModelModal
        model={model}
        providers={providers}
        open
        onOpenChange={() => undefined}
      />
    );

    await userEvent.clear(screen.getByLabelText("Model Rate Limit (RPS)"));
    await userEvent.type(screen.getByLabelText("Model Rate Limit (RPS)"), "3");
    fireEvent.submit(document.querySelector("form") as HTMLFormElement);

    await waitFor(() => expect(updateModel).toHaveBeenCalledWith("model-1", expect.any(FormData)));
    const submitted = vi.mocked(updateModel).mock.calls[0][1];
    expect(submitted.get("rate_limit_rps")).toBe("3");
  });

  it("keeps entered rate limit visible after a failed save", async () => {
    vi.mocked(updateModel).mockRejectedValue(new Error("gateway rejected update"));
    render(
      <EditModelModal
        model={model}
        providers={providers}
        open
        onOpenChange={() => undefined}
      />
    );

    await userEvent.clear(screen.getByLabelText("Model Rate Limit (RPS)"));
    await userEvent.type(screen.getByLabelText("Model Rate Limit (RPS)"), "9");
    fireEvent.submit(document.querySelector("form") as HTMLFormElement);

    await waitFor(() => expect(updateModel).toHaveBeenCalled());
    expect(screen.getByText("gateway rejected update")).toBeInTheDocument();
    expect(screen.getByLabelText("Model Rate Limit (RPS)")).toHaveValue(9);
  });
});
