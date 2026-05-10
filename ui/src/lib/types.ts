export interface APIKey {
  id: string;
  key_hash: string;
  name: string;
  rate_limit_rps: number;
  is_active: boolean;
  expires_at?: string | null;
  metadata?: Record<string, string> | null;
  allowed_models?: string[] | null;
  created_at: string;
  updated_at: string;
}

export type BudgetDuration = "daily" | "weekly" | "monthly";

export interface APIKeyBudget {
  api_key_id: string;
  budget_usd: number;
  spent_usd: number;
  remaining_usd: number;
  alert_threshold_pct?: number | null;
  budget_duration?: BudgetDuration | null;
  budget_reset_at?: string | null;
}

export type APIKeyBudgetState =
  | { status: "configured"; budget: APIKeyBudget }
  | { status: "unbudgeted" }
  | { status: "unavailable"; message: string };

export interface APIKeyBudgetFormValues {
  budget_usd: number;
  alert_threshold_pct?: number;
  budget_duration?: BudgetDuration;
  budget_reset_at?: string;
}

export type BudgetStatusIndicator =
  | "unbudgeted"
  | "within_budget"
  | "near_threshold"
  | "exhausted"
  | "unavailable";

export interface Model {
  id: string;
  provider_id: string;
  name: string;
  alias: string;
  context_window: number;
  rate_limit_rps: number;
  created_at: string;
  updated_at: string;
}

export interface APIKeyFormValues {
  name: string;
  rate_limit_rps: number;
  expires_at: string | null;
  metadata: Record<string, string>;
  allowed_models: string[];
  is_active?: boolean;
}

export type APIKeyMutationResult =
  | { success: true; api_key?: string }
  | { success: false; error: string };

export interface GatewayErrorEnvelope {
  error?: {
    message?: string;
    type?: string;
    code?: string;
  } | string;
  message?: string;
}

export interface Provider {
  id: string;
  name: string;
  base_url: string;
  auth_type: string;
  has_credentials?: boolean;
  created_at?: string;
  updated_at?: string;
}

export type UsageGroupBy = "model" | "provider" | "api_key";
export type AuditStatusFilter = "success" | "error";
export type AnalyticsStateKind =
  | "ready"
  | "empty"
  | "unavailable"
  | "invalid-filter";

export interface UsageSummaryRow {
  group: string;
  requests: number;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  cost_usd: number;
  errors: number;
  avg_latency_ms: number;
}

export interface UsageSummaryResponse {
  group_by: UsageGroupBy;
  from: string;
  to: string;
  data: UsageSummaryRow[];
}

export interface DailyUsagePoint {
  date: string;
  requests: number;
  total_tokens: number;
  cost_usd: number;
}

export interface DailyUsageResponse {
  from: string;
  to: string;
  data: DailyUsagePoint[];
}

export interface AuditRecord {
  id: number;
  request_id: string;
  api_key_id: string;
  model: string;
  provider: string;
  prompt_tokens: number;
  output_tokens: number;
  total_tokens: number;
  cost_usd: number;
  status_code: number;
  latency_ms: number;
  streaming: boolean;
  error_message: string;
  created_at: string;
}

export interface AuditResponse {
  data: AuditRecord[];
  total: number;
  limit: number;
  offset: number;
}

export interface AnalyticsFilters {
  from?: string;
  to?: string;
  group_by: UsageGroupBy;
  api_key_id?: string;
  model?: string;
  provider?: string;
  status?: AuditStatusFilter;
  limit: number;
  offset: number;
}

export interface AnalyticsResult<T> {
  state: AnalyticsStateKind;
  data: T;
  message?: string;
}
