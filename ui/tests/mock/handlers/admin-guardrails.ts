import { Hono } from "hono";
import { getStore } from "../state";

const app = new Hono();

app.get("/", (c) => c.json({ data: getStore().guardrails }));

// PATCH persists within a worker session (in-memory state survives reloads
// because the mock server is a single long-running process per Playwright worker).
app.patch("/:id", async (c) => {
  const store = getStore();
  const idx = store.guardrails.findIndex((g) => g.id === c.req.param("id"));
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  Object.assign(store.guardrails[idx], await c.req.json());
  return c.json(store.guardrails[idx]);
});

export default app;
