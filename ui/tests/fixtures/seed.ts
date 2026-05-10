import { providers, type ProviderFixture } from "./providers";
import { models } from "./models";
import { keys, type KeyFixture } from "./keys";
import { teams } from "./teams";
import { teamBudgets } from "./budgets";
import { rateLimits } from "./rate-limits";
import { guardrails } from "./guardrails";
import { logs } from "./logs";
import { auditEntries } from "./audit";
import type { MockTeam, TeamBudget, MockRateLimitRule, MockGuardrail, MockLogRow } from "../mock/types";
import type { Model, AuditRecord } from "../../src/lib/types";

export interface SeedSnapshot {
  providers: ProviderFixture[];
  models: Model[];
  keys: KeyFixture[];
  teams: MockTeam[];
  teamBudgets: TeamBudget[];
  rateLimits: MockRateLimitRule[];
  guardrails: MockGuardrail[];
  logs: MockLogRow[];
  auditEntries: AuditRecord[];
  // One-time rotate reveals: key_id → full secret string.
  pendingReveal: Record<string, string>;
}

// Returns a fresh deep copy of the seed data.
export function seed(): SeedSnapshot {
  return {
    providers: structuredClone(providers),
    models: structuredClone(models),
    keys: structuredClone(keys),
    teams: structuredClone(teams),
    teamBudgets: structuredClone(teamBudgets),
    rateLimits: structuredClone(rateLimits),
    guardrails: structuredClone(guardrails),
    logs: structuredClone(logs),
    auditEntries: structuredClone(auditEntries),
    pendingReveal: {},
  };
}
