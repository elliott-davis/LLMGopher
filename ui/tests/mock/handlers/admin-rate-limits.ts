import { Hono } from "hono";
import { getStore } from "../state";
import type { MockRateLimitRule } from "../types";

const app = new Hono();

app.get("/", (c) => c.json({ data: getStore().rateLimits }));

app.post("/", async (c) => {
  const body = await c.req.json<Partial<MockRateLimitRule>>();
  const rule: MockRateLimitRule = {
    id: `rl_${Date.now()}`,
    scope: body.scope ?? "model",
    scope_id: body.scope_id ?? "",
    rps: body.rps ?? 100,
    tpm: body.tpm,
    tripped: false,
  };
  getStore().rateLimits.push(rule);
  return c.json(rule, 201);
});

app.patch("/:id", async (c) => {
  const store = getStore();
  const idx = store.rateLimits.findIndex((r) => r.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  Object.assign(store.rateLimits[idx], await c.req.json());
  return c.json(store.rateLimits[idx]);
});

app.delete("/:id", (c) => {
  const store = getStore();
  const idx = store.rateLimits.findIndex((r) => r.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  store.rateLimits.splice(idx, 1);
  return c.body(null, 204);
});

export default app;
