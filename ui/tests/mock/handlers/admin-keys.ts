import { Hono } from "hono";
import { getStore } from "../state";
import type { KeyFixture } from "../../fixtures/keys";

const app = new Hono();

// Real gateway returns APIKey[] (raw array, no wrapper).
// Never expose _budget or key secrets in list responses.
app.get("/", (c) => {
  const rows = getStore().keys.map(({ _budget: _, ...k }) => k);
  return c.json(rows);
});

app.post("/", async (c) => {
  const body = await c.req.json<Partial<KeyFixture> & { name?: string }>();
  if (!body.name) return c.json({ error: { message: "name is required", type: "invalid_request_error" } }, 400);
  const now = new Date().toISOString();
  const secret = `sk-mock-${Date.now()}`;
  const key: KeyFixture = {
    id: `key_${Date.now()}`,
    key_hash: `sha256:mock${Date.now()}`,
    name: body.name,
    rate_limit_rps: body.rate_limit_rps ?? 0,
    is_active: true,
    expires_at: body.expires_at ?? null,
    allowed_models: body.allowed_models ?? null,
    created_at: now,
    updated_at: now,
  };
  getStore().keys.push(key);
  // One-time reveal: full key in creation response only.
  const { _budget: _, ...keyOut } = key;
  return c.json({ ...keyOut, key: secret }, 201);
});

app.patch("/:id", async (c) => {
  const store = getStore();
  const idx = store.keys.findIndex((k) => k.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  Object.assign(store.keys[idx], await c.req.json());
  store.keys[idx].updated_at = new Date().toISOString();
  const { _budget: _, ...keyOut } = store.keys[idx];
  return c.json(keyOut);
});

// Rotate: generate new prefix, return one-time full key.
app.post("/:id/rotate", (c) => {
  const store = getStore();
  const key = store.keys.find((k) => k.id === c.req.param("id"));
  if (!key) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  const newSecret = `sk-mock-rot-${Date.now()}`;
  store.pendingReveal[key.id] = newSecret;
  key.updated_at = new Date().toISOString();
  return c.json({ id: key.id, key: newSecret, prefix: newSecret.slice(-4) });
});

// Budget sub-resource
app.put("/:id/budget", async (c) => {
  const store = getStore();
  const key = store.keys.find((k) => k.id === c.req.param("id"));
  if (!key) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  const body = await c.req.json<{ budget_usd: number; budget_duration?: string; alert_threshold_pct?: number }>();
  key._budget = {
    limit_usd: body.budget_usd,
    usage_usd: 0,
    duration: (body.budget_duration ?? "monthly") as "monthly",
    alert_threshold: (body.alert_threshold_pct ?? 85) / 100,
  };
  return c.json({ api_key_id: key.id, budget_usd: key._budget.limit_usd, spent_usd: 0, remaining_usd: key._budget.limit_usd });
});

app.post("/:id/budget/reset", (c) => {
  const store = getStore();
  const key = store.keys.find((k) => k.id === c.req.param("id"));
  if (!key) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  if (key._budget) key._budget.usage_usd = 0;
  return c.json({ reset: true });
});

app.delete("/:id/budget", (c) => {
  const store = getStore();
  const key = store.keys.find((k) => k.id === c.req.param("id"));
  if (!key) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  delete key._budget;
  return c.body(null, 204);
});

export default app;
