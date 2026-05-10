import { Hono } from "hono";
import { getStore } from "../state";
import type { Model } from "../../../src/lib/types";

const app = new Hono();

// Real gateway returns Model[] (raw array, no wrapper).
app.get("/", (c) => c.json(getStore().models));

app.post("/", async (c) => {
  const body = await c.req.json<Partial<Model>>();
  if (!body.id || !body.provider_id) {
    return c.json({ error: { message: "id and provider_id are required", type: "invalid_request_error" } }, 400);
  }
  const now = new Date().toISOString();
  const model: Model = {
    id: body.id,
    provider_id: body.provider_id,
    name: body.name ?? body.id,
    alias: body.alias ?? body.id,
    context_window: body.context_window ?? 4096,
    rate_limit_rps: body.rate_limit_rps ?? 0,
    created_at: now,
    updated_at: now,
  };
  getStore().models.push(model);
  return c.json(model, 201);
});

app.patch("/:id", async (c) => {
  const store = getStore();
  const idx = store.models.findIndex((m) => m.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  Object.assign(store.models[idx], await c.req.json());
  store.models[idx].updated_at = new Date().toISOString();
  return c.json(store.models[idx]);
});

app.delete("/:id", (c) => {
  const store = getStore();
  const idx = store.models.findIndex((m) => m.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  store.models.splice(idx, 1);
  return c.body(null, 204);
});

export default app;
