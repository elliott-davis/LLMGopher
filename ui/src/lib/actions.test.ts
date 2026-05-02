import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("next/cache", () => ({
  revalidatePath: vi.fn(),
}));

import {
  createAPIKey,
  deleteAPIKey,
  setAPIKeyActiveState,
  updateAPIKey,
  waitForAPIKeyDeletionSync,
} from "@/lib/actions";
import {
  extractGatewayErrorMessage,
  parseAPIKeyFormValues,
  parseAPIKeyMetadata,
} from "@/lib/action-helpers";
import { revalidatePath } from "next/cache";

function formData(values: Record<string, string | string[]>) {
  const form = new FormData();
  for (const [key, value] of Object.entries(values)) {
    if (Array.isArray(value)) {
      value.forEach((item) => form.append(key, item));
    } else {
      form.set(key, value);
    }
  }
  return form;
}

describe("API key lifecycle actions", () => {
  beforeEach(() => {
    vi.mocked(revalidatePath).mockClear();
    vi.stubGlobal("fetch", vi.fn());
  });

  it("parses lifecycle form fields", () => {
    const values = parseAPIKeyFormValues(
      formData({
        name: " Production ",
        rate_limit_rps: "10",
        expires_at: "2026-12-31T23:59",
        metadata: '{"owner":"platform"}',
        allowed_models: ["gpt-4o", "claude-3"],
        is_active: "true",
      })
    );

    expect(values).toMatchObject({
      name: "Production",
      rate_limit_rps: 10,
      metadata: { owner: "platform" },
      allowed_models: ["gpt-4o", "claude-3"],
      is_active: true,
    });
    expect(values.expires_at).toBe("2026-12-31T23:59:00.000Z");
  });

  it("rejects invalid metadata and negative rate limits", () => {
    expect(() => parseAPIKeyMetadata('{"owner": 42}' as unknown as string)).toThrow(
      "Metadata values must be strings"
    );
    expect(() =>
      parseAPIKeyFormValues(formData({ name: "bad", rate_limit_rps: "-1" }))
    ).toThrow("Invalid API key form data");
  });

  it("extracts nested gateway error envelope messages", () => {
    expect(
      extractGatewayErrorMessage(
        { error: { message: "api key not found", type: "invalid_request_error" } },
        "fallback"
      )
    ).toBe("api key not found");
  });

  it("creates API keys with lifecycle payload and returns raw key once", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ api_key: "sk-created" }), { status: 200 })
    );

    const rawKey = await createAPIKey(
      formData({
        name: "prod",
        rate_limit_rps: "5",
        metadata: '{"owner":"platform"}',
        allowed_models: ["gpt-4o"],
      })
    );

    expect(rawKey).toBe("sk-created");
    expect(fetch).toHaveBeenCalledWith(
      "http://gateway:8080/v1/admin/keys",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          name: "prod",
          rate_limit_rps: 5,
          expires_at: null,
          metadata: { owner: "platform" },
          allowed_models: ["gpt-4o"],
        }),
      })
    );
    expect(revalidatePath).toHaveBeenCalledWith("/keys");
  });

  it("updates API keys and surfaces not-found gateway errors", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(
        JSON.stringify({ error: { message: "api key not found" } }),
        { status: 404 }
      )
    );

    await expect(
      updateAPIKey("key-1", formData({ name: "prod", rate_limit_rps: "5" }))
    ).rejects.toThrow("api key not found");
  });

  it("deactivates and reactivates API keys with is_active payloads", async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 200 }));

    await setAPIKeyActiveState("key-1", false);
    await setAPIKeyActiveState("key-1", true);

    expect(fetch).toHaveBeenNthCalledWith(
      1,
      "http://gateway:8080/v1/admin/keys/key-1",
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({ is_active: false }),
      })
    );
    expect(fetch).toHaveBeenNthCalledWith(
      2,
      "http://gateway:8080/v1/admin/keys/key-1",
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({ is_active: true }),
      })
    );
  });

  it("deletes keys and reports deletion sync timeout", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValue(
        new Response(JSON.stringify([{ id: "key-1" }]), { status: 200 })
      );

    await deleteAPIKey("key-1");
    const synced = await waitForAPIKeyDeletionSync("key-1", 1, 1);

    expect(synced).toBe(false);
  });
});
