import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import KeysPage from "@/app/(dashboard)/keys/page";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    refresh: vi.fn(),
  }),
}));

vi.mock("@/lib/actions", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/actions")>()),
  createAPIKey: vi.fn(),
  deleteAPIKey: vi.fn(),
  setAPIKeyActiveState: vi.fn(),
  updateAPIKey: vi.fn(),
  waitForAPIKeyDeletionSync: vi.fn(),
}));

describe("KeysPage", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  it("renders lifecycle fields and stale model labels", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify([
            {
              id: "key-1",
              key_hash: "abc123",
              name: "Production",
              rate_limit_rps: 10,
              is_active: true,
              expires_at: "2026-12-31T23:59:00Z",
              metadata: { owner: "platform" },
              allowed_models: ["legacy-model"],
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ]),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { status: 200 }));

    render(await KeysPage());

    expect(screen.getByText("Production")).toBeInTheDocument();
    expect(screen.getByText("legacy-model (stale)")).toBeInTheDocument();
    expect(screen.getByText("owner: platform")).toBeInTheDocument();
  });

  it("uses the expanded column span for unavailable inventory", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response(null, { status: 503 }))
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { status: 200 }));

    render(await KeysPage());

    expect(
      screen.getByText("Backend unavailable. Try refreshing in a moment.")
    ).toHaveAttribute("colspan", "9");
  });
});
