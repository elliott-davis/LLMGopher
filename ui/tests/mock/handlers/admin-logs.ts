import { Hono } from "hono";
import { getStore } from "../state";

const app = new Hono();

app.get("/", (c) => {
  const status = c.req.query("status"); // "2xx" | "4xx" | "5xx" | "fallback"
  let rows = getStore().logs;
  if (status) {
    rows = rows.filter((r) => {
      if (status === "2xx") return r.status_code >= 200 && r.status_code < 300;
      if (status === "4xx") return r.status_code >= 400 && r.status_code < 500;
      if (status === "5xx") return r.status_code >= 500 && r.status_code < 600;
      if (status === "fallback") return r.provider_chain.length > 1 && r.provider_chain[0].status === "failed";
      return true;
    });
  }
  const limit = Number(c.req.query("limit") ?? 20);
  return c.json({ data: rows.slice(0, limit), next_cursor: null });
});

app.get("/:id", (c) => {
  const row = getStore().logs.find((l) => l.id === c.req.param("id"));
  if (!row) return c.json({ error: { message: "not found", type: "not_found_error" } }, 404);
  return c.json({
    ...row,
    prompt_preview: "What is the capital of France?",
    response_preview: "Paris.",
    trace: row.provider_chain,
    headers: {
      "content-type": "application/json",
      authorization: "Bearer sk-test-key-1",
    },
  });
});

export default app;
