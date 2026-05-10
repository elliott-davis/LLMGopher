import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import ModelsPage from "@/app/(dashboard)/models/page";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    refresh: vi.fn(),
  }),
}));

vi.mock("@/lib/actions", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/actions")>()),
  createModel: vi.fn(),
  deleteModel: vi.fn(),
  updateModel: vi.fn(),
  waitForModelDeletionSync: vi.fn(),
}));

describe("ModelsPage", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  it("renders a dedicated model rate limit column with limited and no-limit states", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify([
            {
              id: "model-1",
              alias: "gpt-4o-limited",
              name: "gpt-4o",
              provider_id: "provider-1",
              context_window: 128000,
              rate_limit_rps: 2,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: "model-2",
              alias: "gpt-4o-unlimited",
              name: "gpt-4o",
              provider_id: "provider-1",
              context_window: 128000,
              rate_limit_rps: 0,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: "model-3",
              alias: "legacy-missing",
              name: "legacy",
              provider_id: "provider-1",
              context_window: 8192,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ]),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { status: 200 }));

    render(await ModelsPage());

    expect(screen.getByRole("columnheader", { name: "Model Rate Limit" })).toBeInTheDocument();
    expect(screen.getByText("2 requests/sec")).toBeInTheDocument();
    expect(screen.getAllByText("No model-level limit")).toHaveLength(2);
    expect(
      screen.getByText(/model-level request limits separate from API key limits/)
    ).toBeInTheDocument();
  });

  it("uses the expanded column span for unavailable and empty inventory", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response(null, { status: 503 }))
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { status: 200 }));

    const { rerender } = render(await ModelsPage());

    expect(
      screen.getByText("Backend unavailable. Try refreshing in a moment.")
    ).toHaveAttribute("colspan", "7");

    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { status: 200 }));

    rerender(await ModelsPage());

    expect(screen.getByText("No models are currently loaded.")).toHaveAttribute(
      "colspan",
      "7"
    );
  });
});
