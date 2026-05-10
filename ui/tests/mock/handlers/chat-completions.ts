import { Hono } from "hono";
import { getStore } from "../state";

const app = new Hono();

// Only budget cap-enforcement is implemented; all other chat traffic returns 501.
app.post("/", async (c) => {
  const auth = c.req.header("authorization") ?? "";
  // Accept "Bearer <key-id>" where key-id is the internal mock id.
  const keyId = auth.replace(/^Bearer\s+/i, "").trim();
  const key = getStore().keys.find((k) => k.id === keyId);

  if (key?._budget && key._budget.usage_usd >= key._budget.limit_usd) {
    return c.json(
      { error: { message: "budget exceeded", type: "rate_limit_error", code: "budget_exceeded" } },
      429,
      { "x-llmgopher-reason": "budget_exceeded" },
    );
  }

  return c.json({ error: { message: "not implemented by mock", type: "internal_error" } }, 501);
});

export default app;
