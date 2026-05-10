// Re-exports from src/lib/types.ts so mock handlers share the same shapes.
export type {
  APIKey,
  APIKeyBudget,
  Model,
  Provider,
  AuditRecord,
} from "../../src/lib/types";

// Extended fixture types (with extra mock-only fields)
export type { ProviderFixture } from "../fixtures/providers";
export type { KeyFixture } from "../fixtures/keys";

// Types not yet in src/lib/types.ts — will migrate when the UI surfaces ship.

export interface MockTeam {
  id: string;
  display_name: string;
  member_count: number;
  budget_utilization: number;
}

export interface TeamBudget {
  scope: "team";
  scope_id: string;
  limit_usd: string;
  usage_usd: string;
  duration: "daily" | "weekly" | "monthly";
  alert_threshold: number;
}

export interface MockRateLimitRule {
  id: string;
  scope: "key" | "model" | "team";
  scope_id: string;
  rps: number;
  tpm?: number;
  tripped: boolean;
}

export interface MockGuardrail {
  id: string;
  display_name: string;
  enabled: boolean;
}

export type { RoutePolicy, RouteTarget, SettingCard, SettingField } from '../../src/lib/admin-surface-contracts';

export interface LogProviderStage {
  provider_id: string;
  status: "ok" | "failed";
  latency_ms: number;
}

export interface MockLogRow {
  id: string;
  request_id: string;
  timestamp: string;
  method: string;
  path: string;
  status_code: number;
  latency_ms: number;
  api_key_id: string;
  model: string;
  provider_chain: LogProviderStage[];
}
