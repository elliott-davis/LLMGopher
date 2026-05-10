import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("next/cache", () => ({
  revalidatePath: vi.fn(),
}));

import {
  createAPIKey,
  createModel,
  deleteAPIKeyBudget,
  deleteAPIKey,
  fetchAPIKeyBudget,
  resetAPIKeyBudget,
  setAPIKeyActiveState,
  upsertAPIKeyBudget,
  updateAPIKey,
  updateModel,
  waitForAPIKeyDeletionSync,
} from "@/lib/actions";
import {
  extractGatewayErrorMessage,
  parseAPIKeyFormValues,
  parseAPIKeyMetadata,
  parseModelFormValues,
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

describe("model actions", () => {
  beforeEach(() => {
    vi.mocked(revalidatePath).mockClear();
    vi.stubGlobal("fetch", vi.fn());
  });

  it("parses model form fields with a default no-limit rate policy", () => {
    expect(
      parseModelFormValues(
        formData({
          alias: " gpt-4o ",
          name: " gpt-4o-2024-11-20 ",
          provider_id: " provider-1 ",
          context_window: "128000",
        })
      )
    ).toEqual({
      alias: "gpt-4o",
      name: "gpt-4o-2024-11-20",
      provider_id: "provider-1",
      context_window: 128000,
      rate_limit_rps: 0,
    });
  });

  it("rejects negative and non-integer model rate limits", () => {
    const valid = {
      alias: "gpt-4o",
      name: "gpt-4o-2024-11-20",
      provider_id: "provider-1",
      context_window: "128000",
    };

    expect(() =>
      parseModelFormValues(formData({ ...valid, rate_limit_rps: "-1" }))
    ).toThrow("Model rate limit must be a non-negative whole number");
    expect(() =>
      parseModelFormValues(formData({ ...valid, rate_limit_rps: "1.5" }))
    ).toThrow("Model rate limit must be a non-negative whole number");
  });

  it("creates models with rate_limit_rps in the gateway payload", async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 201 }));

    await createModel(
      formData({
        alias: "gpt-4o",
        name: "gpt-4o-2024-11-20",
        provider_id: "provider-1",
        context_window: "128000",
        rate_limit_rps: "25",
      })
    );

    expect(fetch).toHaveBeenCalledWith(
      "http://gateway:8080/v1/admin/models",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          alias: "gpt-4o",
          name: "gpt-4o-2024-11-20",
          provider_id: "provider-1",
          context_window: 128000,
          rate_limit_rps: 25,
        }),
      })
    );
    expect(revalidatePath).toHaveBeenCalledWith("/models");
  });

  it("updates models with rate_limit_rps in the gateway payload", async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 200 }));

    await updateModel(
      "model-1",
      formData({
        alias: "gpt-4o",
        name: "gpt-4o-2024-11-20",
        provider_id: "provider-1",
        context_window: "128000",
        rate_limit_rps: "0",
      })
    );

    expect(fetch).toHaveBeenCalledWith(
      "http://gateway:8080/v1/admin/models/model-1",
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({
          alias: "gpt-4o",
          name: "gpt-4o-2024-11-20",
          provider_id: "provider-1",
          context_window: 128000,
          rate_limit_rps: 0,
        }),
      })
    );
  });

  it("surfaces gateway envelope messages for model create and update failures", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({ error: { message: "model rate limit rejected" } }),
          { status: 400 }
        )
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({ error: { message: "model rate limit rejected" } }),
          { status: 400 }
        )
      );

    await expect(
      createModel(
        formData({
          alias: "gpt-4o",
          name: "gpt-4o-2024-11-20",
          provider_id: "provider-1",
          context_window: "128000",
          rate_limit_rps: "3",
        })
      )
    ).rejects.toThrow("model rate limit rejected");

    await expect(
      updateModel(
        "model-1",
        formData({
          alias: "gpt-4o",
          name: "gpt-4o-2024-11-20",
          provider_id: "provider-1",
          context_window: "128000",
          rate_limit_rps: "3",
        })
      )
    ).rejects.toThrow("model rate limit rejected");
  });
});

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
      .mockImplementation(() =>
        Promise.resolve(
          new Response(JSON.stringify([{ id: "key-1" }]), { status: 200 })
        )
      );

    await deleteAPIKey("key-1");
    const synced = await waitForAPIKeyDeletionSync("key-1", 1, 1);

    expect(synced).toBe(false);
  });
});

describe("API key budget actions", () => {
  beforeEach(() => {
    vi.mocked(revalidatePath).mockClear();
    vi.stubGlobal("fetch", vi.fn());
    process.env.LLMGOPHER_UI_ADMIN_API_KEY = "sk-test-key-1";
  });

  it("fetches configured budget state and maps 404 to unbudgeted", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            api_key_id: "key-1",
            budget_usd: 100,
            spent_usd: 10,
            remaining_usd: 90,
          }),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(new Response(null, { status: 404 }));

    await expect(fetchAPIKeyBudget("key-1")).resolves.toMatchObject({
      status: "configured",
    });
    await expect(fetchAPIKeyBudget("key-2")).resolves.toEqual({ status: "unbudgeted" });
    expect(fetch).toHaveBeenCalledWith(
      "http://gateway:8080/v1/admin/keys/key-1/budget",
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer sk-test-key-1",
        }),
      })
    );
  });

  it("returns unavailable state for missing token and auth failures", async () => {
    delete process.env.LLMGOPHER_UI_ADMIN_API_KEY;
    await expect(fetchAPIKeyBudget("key-1")).resolves.toEqual(
      expect.objectContaining({ status: "unavailable" })
    );

    process.env.LLMGOPHER_UI_ADMIN_API_KEY = "sk-test-key-1";
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 401 }));
    await expect(fetchAPIKeyBudget("key-1")).resolves.toEqual(
      expect.objectContaining({ status: "unavailable" })
    );
  });

  it("upserts, resets, and deletes budget with revalidation", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            api_key_id: "key-1",
            budget_usd: 100,
            spent_usd: 20,
            remaining_usd: 80,
            alert_threshold_pct: 80,
          }),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            api_key_id: "key-1",
            budget_usd: 100,
            spent_usd: 0,
            remaining_usd: 100,
          }),
          { status: 200 }
        )
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    const form = new FormData();
    form.set("budget_usd", "100");
    form.set("alert_threshold_pct", "80");

    await expect(upsertAPIKeyBudget("key-1", form)).resolves.toMatchObject({
      api_key_id: "key-1",
      budget_usd: 100,
    });
    await expect(resetAPIKeyBudget("key-1")).resolves.toMatchObject({
      spent_usd: 0,
    });
    await expect(deleteAPIKeyBudget("key-1")).resolves.toBeUndefined();
    expect(revalidatePath).toHaveBeenCalledWith("/keys");
  });

  it("surfaces gateway envelope messages for budget mutations", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ error: { message: "budget invalid" } }), {
        status: 400,
      })
    );
    const form = new FormData();
    form.set("budget_usd", "100");
    await expect(upsertAPIKeyBudget("key-1", form)).rejects.toThrow("budget invalid");
  });
});
