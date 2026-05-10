export interface ProviderStage {
  provider_id: string;
  provider_name?: string;
  status: 'ok' | 'success' | 'failed' | 'skipped' | 'unavailable';
  latency_ms?: number;
  error_summary?: string;
}

export interface LogRow {
  id: string;
  request_id: string;
  timestamp: string;
  method: string;
  path: string;
  status_code: number;
  latency_ms: number;
  api_key_id: string;
  model: string;
  provider_chain: ProviderStage[];
}

export interface LogDetail extends LogRow {
  prompt_preview: string;
  response_preview: string;
  headers: Record<string, string>;
  trace: ProviderStage[];
}

export interface LogsFilter {
  status?: 'all' | '2xx' | '4xx' | '5xx' | 'fallback';
  page?: number;
  limit?: number;
}

export interface LogsListResponse {
  data: LogRow[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

export interface AuditFilter {
  actor?: string;
  action?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}

export interface AuditListResponse {
  data: import('./types').AuditRecord[];
  total: number;
  limit: number;
  offset: number;
}

export interface RouteTarget {
  provider_id: string;
  provider_name: string;
  weight: number;
  order?: number;
  health_state?: string;
  latency_ms?: number;
}

export interface RoutePolicy {
  id: string;
  model_alias: string;
  strategy: 'single' | 'fallback' | 'weighted' | 'latency';
  enabled: boolean;
  targets: RouteTarget[];
  primary_provider_id?: string;
  fallback_provider_ids?: string[];
  health_summary?: string;
}

export interface GuardrailRule {
  id: string;
  display_name: string;
  enabled: boolean;
  category?: string;
  description?: string;
  provider_label?: string;
  last_updated_at?: string;
}

export interface Team {
  id: string;
  display_name: string;
  member_count: number;
  budget_utilization: number;
  budget_health: 'ok' | 'near_cap' | 'over_cap';
}

export interface BudgetPolicy {
  scope: 'team' | 'key';
  scope_id: string;
  display_name?: string;
  limit_usd: number;
  usage_usd: number;
  duration: 'daily' | 'weekly' | 'monthly';
  alert_threshold: number;
  hard_cap_state: 'ok' | 'near_cap' | 'over_cap';
}

export interface RateLimitRule {
  id: string;
  scope: 'model' | 'key' | 'team';
  scope_id: string;
  rps: number;
  tpm?: number;
  tripped: boolean;
}

export interface SettingField {
  id: string;
  label: string;
  value: string;
  input_type: string;
  read_only: boolean;
  validation_message?: string;
}

export interface SettingCard {
  id: 'gateway-profile' | 'security' | 'notifications' | 'display';
  title: string;
  description: string;
  availability: 'read_only' | 'editable' | 'unavailable';
  fields: SettingField[];
  save_capability: boolean;
  last_saved_at?: string;
}

export type AdminSurfaceUnavailableState = { kind: 'unavailable'; message: string };
export type AdminSurfaceLoadingState = { kind: 'loading' };
export type AdminSurfaceEmptyState = { kind: 'empty' };
