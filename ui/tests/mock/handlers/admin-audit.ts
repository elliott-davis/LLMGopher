import { Hono } from "hono";
import { getStore } from "../state";

const app = new Hono();

// Returns AuditResponse: { data: AuditRecord[], total, limit, offset }
app.get("/", (c) => {
  const { actor, action, from, to } = c.req.query();
  let entries = getStore().auditEntries;
  if (actor) entries = entries.filter((e) => e.api_key_id === actor);
  if (action) entries = entries.filter((e) => e.model === action);
  if (from)  entries = entries.filter((e) => e.created_at >= from);
  if (to)    entries = entries.filter((e) => e.created_at <= to);
  const limit = Number(c.req.query("limit") ?? 25);
  const offset = Number(c.req.query("offset") ?? 0);
  const page = entries.slice(offset, offset + limit);
  return c.json({ data: page, total: entries.length, limit, offset });
});

export default app;
