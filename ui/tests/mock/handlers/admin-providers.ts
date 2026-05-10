import { Hono } from "hono";
import { getStore } from "../state";
import type { ProviderFixture } from "../../fixtures/providers";

const app = new Hono();

// Real gateway returns Provider[] (raw array, no wrapper).
app.get("/", (c) => c.json(getStore().providers));

app.post("/", async (c) => {
  // The real server action sends FormData (multipart); JSON also accepted for direct API use.
  let name: string | undefined;
  let base_url: string | undefined;
  let auth_type: string | undefined;
  let kind: string | undefined;

  const ct = c.req.header("content-type") ?? "";
  if (ct.includes("multipart/form-data") || ct.includes("application/x-www-form-urlencoded")) {
    const fd = await c.req.formData();
    name = fd.get("name")?.toString();
    base_url = fd.get("base_url")?.toString();
    auth_type = fd.get("auth_type")?.toString();
    kind = fd.get("kind")?.toString();
  } else {
    const body = await c.req.json<Partial<ProviderFixture>>();
    name = body.name;
    base_url = body.base_url;
    auth_type = body.auth_type;
    kind = body.kind;
  }

  if (!name) {
    return c.json({ error: { message: "name is required", type: "invalid_request_error", code: "invalid_request" } }, 400);
  }
  if (base_url) {
    try { new URL(base_url); } catch {
      return c.json({ error: { message: "invalid base_url", type: "invalid_request_error", code: "invalid_url" } }, 400);
    }
  }
  const now = new Date().toISOString();
  const provider: ProviderFixture = {
    id: `prov_${Date.now()}`,
    kind: (kind ?? "generic") as ProviderFixture["kind"],
    name,
    base_url: base_url ?? "",
    auth_type: auth_type ?? "api_key",
    has_credentials: false,
    health: "healthy",
    created_at: now,
    updated_at: now,
  };
  getStore().providers.push(provider);
  return c.json(provider, 201);
});

app.patch("/:id", async (c) => {
  const store = getStore();
  const idx = store.providers.findIndex((p) => p.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  Object.assign(store.providers[idx], await c.req.json());
  return c.json(store.providers[idx]);
});

app.delete("/:id", (c) => {
  const store = getStore();
  const idx = store.providers.findIndex((p) => p.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  store.providers.splice(idx, 1);
  return c.body(null, 204);
});

export default app;
