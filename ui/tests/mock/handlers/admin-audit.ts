import { Hono } from "hono";
import { getStore } from "../state";

const app = new Hono();

// Outcome derivation mirroring the Go backend (deriveOutcome in internal/api/admin.go).
function deriveOutcome(statusCode: number, errorMessage: string): string {
  if (statusCode < 400) return "success";
  if (statusCode === 401 || statusCode === 403) return "unauthorized";
  if (statusCode === 429) {
    return errorMessage?.toLowerCase().includes("budget")
      ? "budget_denied"
      : "rate_limited";
  }
  if (statusCode >= 400 && statusCode < 500) return "client_error";
  return "failure";
}

// Reference summary: emit missing entries for empty actor_id, model, or provider.
function buildReferenceSummary(
  entry: Record<string, unknown>
): Array<{ field: string; original_id: string; state: string }> | undefined {
  const refs: Array<{ field: string; original_id: string; state: string }> = [];
  if (!entry.api_key_id) {
    refs.push({ field: "actor_id", original_id: "", state: "missing" });
  }
  if (!entry.model) {
    refs.push({ field: "model", original_id: "", state: "missing" });
  }
  if (!entry.provider) {
    refs.push({ field: "provider", original_id: "", state: "missing" });
  }
  return refs.length > 0 ? refs : undefined;
}

// Returns AuditResponse: { data: AuditRecord[], total, limit, offset, page, has_more }
// Accepts: actor (preferred), api_key_id (deprecated alias), action, outcome, model, provider,
//          status (legacy), from, to, limit, offset.
app.get("/", (c) => {
  const params = c.req.query();
  const actor = params.actor ?? params.api_key_id ?? "";
  const action = params.action ?? "";
  const outcome = params.outcome ?? "";
  const status = params.status ?? "";
  const from = params.from ?? "";
  const to = params.to ?? "";

  let entries = getStore().auditEntries as Array<Record<string, unknown>>;

  if (actor) entries = entries.filter((e) => e.api_key_id === actor);

  if (action.startsWith("request:")) {
    const modelPart = action.slice("request:".length);
    if (modelPart) {
      // exact selector: action=request:gpt-4o → filter by model
      entries = entries.filter((e) => e.model === modelPart);
    }
    // family selector: action=request: → all rows with a model (no additional filter)
  }

  if (outcome) {
    entries = entries.filter(
      (e) =>
        deriveOutcome(
          Number(e.status_code),
          String(e.error_message ?? "")
        ) === outcome
    );
  } else if (status === "success") {
    entries = entries.filter((e) => Number(e.status_code) < 400);
  } else if (status === "error") {
    entries = entries.filter((e) => Number(e.status_code) >= 400);
  }

  if (from) entries = entries.filter((e) => String(e.created_at) >= from);
  if (to) entries = entries.filter((e) => String(e.created_at) <= to);

  const limit = Math.min(Number(params.limit ?? 25), 1000);
  const offset = Math.max(Number(params.offset ?? 0), 0);
  const total = entries.length;
  const pageData = entries.slice(offset, offset + limit);
  const page = Math.floor(offset / limit) + 1;
  const has_more = offset + pageData.length < total;

  // Enrich each row with UI-aligned fields.
  const data = pageData.map((e) => {
    const referenceSummary = buildReferenceSummary(e);
    const row: Record<string, unknown> = {
      ...e,
      actor_id: e.api_key_id ?? "",
      action: `request:${String(e.model ?? "")}`,
      outcome: deriveOutcome(Number(e.status_code), String(e.error_message ?? "")),
    };
    if (referenceSummary) row.reference_summary = referenceSummary;
    return row;
  });

  return c.json({ data, total, limit, offset, page, has_more });
});

export default app;
