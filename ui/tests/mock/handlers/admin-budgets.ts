import { Hono } from "hono";
import { getStore } from "../state";

const app = new Hono();

app.get("/", (c) => c.json({ data: getStore().teamBudgets }));

app.patch("/:scope/:scope_id", async (c) => {
  const store = getStore();
  const { scope, scope_id } = c.req.param();
  const idx = store.teamBudgets.findIndex((b) => b.scope === scope && b.scope_id === scope_id);
  if (idx === -1) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  Object.assign(store.teamBudgets[idx], await c.req.json());
  return c.json(store.teamBudgets[idx]);
});

export default app;
