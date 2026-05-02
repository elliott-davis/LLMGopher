import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import APIKeyRowActions from "@/components/APIKeyRowActions";
import { APIKey } from "@/lib/types";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    refresh: vi.fn(),
  }),
}));

vi.mock("@/lib/actions", () => ({
  deleteAPIKey: vi.fn(),
  setAPIKeyActiveState: vi.fn(),
  waitForAPIKeyDeletionSync: vi.fn(),
}));

const apiKey: APIKey = {
  id: "key-1",
  key_hash: "abc123",
  name: "Production",
  rate_limit_rps: 10,
  is_active: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("APIKeyRowActions", () => {
  it("renders a row actions trigger without raw key material", () => {
    render(<APIKeyRowActions apiKey={apiKey} models={[]} />);

    expect(screen.getByRole("button", { name: "Open actions" })).toBeInTheDocument();
    expect(screen.queryByText(/^sk-/)).not.toBeInTheDocument();
  });

  it("keeps inactive rows using the same action entry point", () => {
    render(
      <APIKeyRowActions apiKey={{ ...apiKey, is_active: false }} models={[]} />
    );

    expect(screen.getByRole("button", { name: "Open actions" })).toBeEnabled();
  });
});
