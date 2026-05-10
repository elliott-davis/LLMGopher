import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import APIKeyLifecycleForm from "@/components/APIKeyLifecycleForm";
import { APIKey, Model } from "@/lib/types";

const models: Model[] = [
  {
    id: "model-1",
    alias: "gpt-4o",
    name: "gpt-4o-2024-11-20",
    provider_id: "openai",
    context_window: 128000,
    rate_limit_rps: 0,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const apiKey: APIKey = {
  id: "key-1",
  key_hash: "abc123",
  name: "Production",
  rate_limit_rps: 10,
  is_active: true,
  expires_at: "2026-12-31T23:59:00Z",
  metadata: { owner: "platform" },
  allowed_models: ["gpt-4o", "legacy-model"],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("APIKeyLifecycleForm", () => {
  it("renders editable lifecycle fields with existing values", () => {
    render(<APIKeyLifecycleForm idPrefix="edit-key" models={models} apiKey={apiKey} />);

    expect(screen.getByLabelText("Name")).toHaveValue("Production");
    expect(screen.getByLabelText("Rate Limit (RPS)")).toHaveValue(10);
    expect(screen.getByLabelText("Metadata JSON")).toHaveValue(
      JSON.stringify({ owner: "platform" }, null, 2)
    );
    expect(screen.getByLabelText(/gpt-4o/)).toBeChecked();
    expect(screen.getByLabelText(/legacy-model/)).toBeChecked();
  });

  it("explains unrestricted access when no models are selected", () => {
    render(<APIKeyLifecycleForm idPrefix="new-key" models={[]} />);

    expect(
      screen.getByText("No models are configured. This key will be unrestricted.")
    ).toBeInTheDocument();
  });
});
