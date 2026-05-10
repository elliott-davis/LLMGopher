import { serve } from "@hono/node-server";
import { Hono } from "hono";
import { reset } from "./state";
import providers from "./handlers/admin-providers";
import models from "./handlers/admin-models";
import keys from "./handlers/admin-keys";
import budgets from "./handlers/admin-budgets";
import teams from "./handlers/admin-teams";
import rateLimits from "./handlers/admin-rate-limits";
import guardrails from "./handlers/admin-guardrails";
import logs from "./handlers/admin-logs";
import audit from "./handlers/admin-audit";
import chatCompletions from "./handlers/chat-completions";
import usage from "./handlers/admin-usage";
import routes from "./handlers/admin-routes";
import settings from "./handlers/admin-settings";
import { MOCK_PORT } from "../support/mock-port";

const app = new Hono();

// Health check used by Playwright webServer to detect readiness.
app.get("/healthz", (c) => c.json({ ok: true }));

// Admin reset endpoint — Playwright calls this between test files.
app.post("/__reset", (c) => {
  reset();
  return c.json({ ok: true });
});

// Error/latency simulation knobs (from contracts/README.md).
app.use("*", async (c, next) => {
  const errorType = c.req.header("x-mock-error");
  if (errorType) {
    const status = errorType === "rate_limit_error" ? 429
      : errorType === "not_found_error" ? 404
      : errorType === "authentication_error" ? 401
      : 400;
    return c.json({ error: { message: `simulated ${errorType}`, type: errorType } }, status as Parameters<typeof c.json>[1]);
  }
  const latency = Number(c.req.query("__mock_latency_ms") ?? 0);
  if (latency > 0) await new Promise((r) => setTimeout(r, latency));
  return next();
});

// Admin API routes
app.route("/v1/admin/providers",   providers);
app.route("/v1/admin/models",      models);
app.route("/v1/admin/keys",        keys);
app.route("/v1/admin/budgets",     budgets);
app.route("/v1/admin/teams",       teams);
app.route("/v1/admin/rate-limits", rateLimits);
app.route("/v1/admin/guardrails",  guardrails);
app.route("/v1/admin/logs",        logs);
app.route("/v1/admin/audit",       audit);
app.route("/v1/admin/usage",       usage);
app.route("/v1/admin/routes",      routes);
app.route("/v1/admin/settings",    settings);
app.route("/v1/chat/completions",  chatCompletions);

// Start server when run directly: node tests/mock/server.ts
const server = serve({ fetch: app.fetch, port: MOCK_PORT }, () => {
  console.log(`Mock gateway listening on http://127.0.0.1:${MOCK_PORT}`);
});

export { app, server };
