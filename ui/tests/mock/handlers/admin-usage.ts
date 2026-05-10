import { Hono } from "hono";

const app = new Hono();

// Returns UsageSummaryResponse: empty data so overview renders gracefully.
app.get("/", (c) => {
  const group_by = c.req.query("group_by") ?? "model";
  const from = c.req.query("from") ?? new Date().toISOString();
  const to = c.req.query("to") ?? new Date().toISOString();
  return c.json({ group_by, from, to, data: [] });
});

// Returns DailyUsageResponse: empty data.
app.get("/daily", (c) => {
  const from = c.req.query("from") ?? new Date().toISOString();
  const to = c.req.query("to") ?? new Date().toISOString();
  return c.json({ from, to, data: [] });
});

export default app;
