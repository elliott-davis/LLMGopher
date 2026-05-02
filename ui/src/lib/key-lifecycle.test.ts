import { describe, expect, it } from "vitest";

import {
  buildModelAllowlistOptions,
  describeModelAllowlist,
  getModelIdentifier,
} from "@/lib/key-lifecycle";
import { Model } from "@/lib/types";

const models: Model[] = [
  {
    id: "model-1",
    alias: "gpt-4o",
    name: "gpt-4o-2024-11-20",
    provider_id: "openai",
    context_window: 128000,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

describe("key lifecycle model utilities", () => {
  it("uses aliases as submitted model identifiers", () => {
    expect(getModelIdentifier(models[0])).toBe("gpt-4o");
  });

  it("preserves stale selected identifiers", () => {
    const options = buildModelAllowlistOptions(models, ["gpt-4o", "legacy-model"]);

    expect(options).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ id: "legacy-model", isAvailable: false }),
        expect.objectContaining({ id: "gpt-4o", isAvailable: true }),
      ])
    );
  });

  it("describes empty allowlists as unrestricted", () => {
    expect(describeModelAllowlist([], models)).toBe("Unrestricted");
  });
});
