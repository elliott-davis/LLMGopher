import type { TeamBudget } from "../mock/types";

// Team-level budgets. Key-level budgets are embedded in keys.ts.
export const teamBudgets: TeamBudget[] = [
  {
    scope: "team",
    scope_id: "team_research",
    limit_usd: "1000.00",
    usage_usd: "860.00",
    duration: "monthly",
    alert_threshold: 0.85,
  },
  {
    scope: "team",
    scope_id: "team_platform",
    limit_usd: "2000.00",
    usage_usd: "800.00",
    duration: "monthly",
    alert_threshold: 0.85,
  },
];
