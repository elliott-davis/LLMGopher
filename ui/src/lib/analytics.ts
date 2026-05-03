import {
  AnalyticsFilters,
  AnalyticsResult,
  AuditResponse,
  AuditStatusFilter,
  DailyUsageResponse,
  UsageGroupBy,
  UsageSummaryResponse,
} from "@/lib/types";

const GATEWAY_BASE = "http://gateway:8080";
const USAGE_ENDPOINT = `${GATEWAY_BASE}/v1/admin/usage`;
const DAILY_USAGE_ENDPOINT = `${GATEWAY_BASE}/v1/admin/usage/daily`;
const AUDIT_ENDPOINT = `${GATEWAY_BASE}/v1/admin/audit`;
const DEFAULT_LIMIT = 25;
const MAX_LIMIT = 1000;
const AUTH_UNAVAILABLE_MESSAGE =
  "Analytics are unavailable. Set LLMGOPHER_UI_ADMIN_API_KEY for the UI service.";

type SearchParamSource =
  | URLSearchParams
  | Record<string, string | string[] | undefined>;

function readParam(source: SearchParamSource, key: string): string {
  if (source instanceof URLSearchParams) {
    return source.get(key)?.trim() ?? "";
  }
  const value = source[key];
  if (Array.isArray(value)) {
    return String(value[0] ?? "").trim();
  }
  return String(value ?? "").trim();
}

function parseOptionalISO(value: string, fieldName: string): string {
  if (!value) {
    return "";
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    throw new Error(`${fieldName} must be a valid ISO 8601 timestamp`);
  }
  return parsed.toISOString();
}

function parsePositiveInt(value: string, fallback: number): number {
  if (!value) {
    return fallback;
  }
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error("limit must be a positive integer");
  }
  return Math.min(parsed, MAX_LIMIT);
}

function parseNonNegativeInt(value: string, fallback: number): number {
  if (!value) {
    return fallback;
  }
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed < 0) {
    throw new Error("offset must be a non-negative integer");
  }
  return parsed;
}

function parseUsageGroupBy(value: string): UsageGroupBy {
  if (!value) {
    return "model";
  }
  if (value === "model" || value === "provider" || value === "api_key") {
    return value;
  }
  throw new Error("group_by must be one of: model, provider, api_key");
}

function parseAuditStatus(value: string): AuditStatusFilter | undefined {
  if (!value) {
    return undefined;
  }
  if (value === "success" || value === "error") {
    return value;
  }
  throw new Error("status must be one of: success, error");
}

function filterIdentityKey(filters: AnalyticsFilters): string {
  return [
    filters.from ?? "",
    filters.to ?? "",
    filters.group_by,
    filters.api_key_id ?? "",
    filters.model ?? "",
    filters.provider ?? "",
    filters.status ?? "",
  ].join("|");
}

export function parseAnalyticsFilters(source: SearchParamSource): AnalyticsFilters {
  const from = parseOptionalISO(readParam(source, "from"), "from");
  const to = parseOptionalISO(readParam(source, "to"), "to");
  if (from && to && from > to) {
    throw new Error("from must be before or equal to to");
  }

  const filters: AnalyticsFilters = {
    from: from || undefined,
    to: to || undefined,
    group_by: parseUsageGroupBy(readParam(source, "group_by")),
    api_key_id: readParam(source, "api_key_id") || undefined,
    model: readParam(source, "model") || undefined,
    provider: readParam(source, "provider") || undefined,
    status: parseAuditStatus(readParam(source, "status")),
    limit: parsePositiveInt(readParam(source, "limit"), DEFAULT_LIMIT),
    offset: parseNonNegativeInt(readParam(source, "offset"), 0),
  };

  const hasPreviousSnapshot = [
    "prev_from",
    "prev_to",
    "prev_group_by",
    "prev_api_key_id",
    "prev_model",
    "prev_provider",
    "prev_status",
  ].some((key) => readParam(source, key) !== "");

  const previous: AnalyticsFilters = {
    from: parseOptionalISO(readParam(source, "prev_from"), "prev_from") || undefined,
    to: parseOptionalISO(readParam(source, "prev_to"), "prev_to") || undefined,
    group_by: parseUsageGroupBy(readParam(source, "prev_group_by") || filters.group_by),
    api_key_id: readParam(source, "prev_api_key_id") || undefined,
    model: readParam(source, "prev_model") || undefined,
    provider: readParam(source, "prev_provider") || undefined,
    status: parseAuditStatus(readParam(source, "prev_status")),
    limit: filters.limit,
    offset: filters.offset,
  };

  if (hasPreviousSnapshot && filterIdentityKey(previous) !== filterIdentityKey(filters)) {
    filters.offset = 0;
  }

  return filters;
}

export function toFilterQueryParams(filters: AnalyticsFilters): URLSearchParams {
  const params = new URLSearchParams();
  params.set("group_by", filters.group_by);
  params.set("limit", String(filters.limit));
  params.set("offset", String(filters.offset));
  if (filters.from) params.set("from", filters.from);
  if (filters.to) params.set("to", filters.to);
  if (filters.api_key_id) params.set("api_key_id", filters.api_key_id);
  if (filters.model) params.set("model", filters.model);
  if (filters.provider) params.set("provider", filters.provider);
  if (filters.status) params.set("status", filters.status);
  return params;
}

function buildUsageQuery(filters: AnalyticsFilters): URLSearchParams {
  const params = new URLSearchParams();
  params.set("group_by", filters.group_by);
  if (filters.from) params.set("from", filters.from);
  if (filters.to) params.set("to", filters.to);
  if (filters.api_key_id) params.set("api_key_id", filters.api_key_id);
  if (filters.model) params.set("model", filters.model);
  return params;
}

function buildDailyUsageQuery(filters: AnalyticsFilters): URLSearchParams {
  const params = new URLSearchParams();
  if (filters.from) params.set("from", filters.from);
  if (filters.to) params.set("to", filters.to);
  if (filters.api_key_id) params.set("api_key_id", filters.api_key_id);
  if (filters.model) params.set("model", filters.model);
  return params;
}

function buildAuditQuery(filters: AnalyticsFilters): URLSearchParams {
  const params = new URLSearchParams();
  params.set("limit", String(filters.limit));
  params.set("offset", String(filters.offset));
  if (filters.from) params.set("from", filters.from);
  if (filters.to) params.set("to", filters.to);
  if (filters.api_key_id) params.set("api_key_id", filters.api_key_id);
  if (filters.model) params.set("model", filters.model);
  if (filters.provider) params.set("provider", filters.provider);
  if (filters.status) params.set("status", filters.status);
  return params;
}

function toFloat(value: unknown, fallback = 0): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function toInt(value: unknown, fallback = 0): number {
  const parsed = Number(value);
  return Number.isInteger(parsed) ? parsed : fallback;
}

function parseUsageSummary(payload: unknown): UsageSummaryResponse {
  const body = payload as Partial<UsageSummaryResponse>;
  return {
    group_by: parseUsageGroupBy(String(body.group_by ?? "model")),
    from: String(body.from ?? ""),
    to: String(body.to ?? ""),
    data: Array.isArray(body.data)
      ? body.data.map((row) => ({
          group: String(row.group ?? "unknown"),
          requests: toInt(row.requests),
          prompt_tokens: toInt(row.prompt_tokens),
          completion_tokens: toInt(row.completion_tokens),
          total_tokens: toInt(row.total_tokens),
          cost_usd: toFloat(row.cost_usd),
          errors: toInt(row.errors),
          avg_latency_ms: toFloat(row.avg_latency_ms),
        }))
      : [],
  };
}

function parseDailyUsage(payload: unknown): DailyUsageResponse {
  const body = payload as Partial<DailyUsageResponse>;
  return {
    from: String(body.from ?? ""),
    to: String(body.to ?? ""),
    data: Array.isArray(body.data)
      ? body.data.map((point) => ({
          date: String(point.date ?? ""),
          requests: toInt(point.requests),
          total_tokens: toInt(point.total_tokens),
          cost_usd: toFloat(point.cost_usd),
        }))
      : [],
  };
}

function parseAudit(payload: unknown): AuditResponse {
  const body = payload as Partial<AuditResponse>;
  return {
    data: Array.isArray(body.data)
      ? body.data.map((row) => ({
          id: toInt(row.id),
          request_id: String(row.request_id ?? ""),
          api_key_id: String(row.api_key_id ?? ""),
          model: String(row.model ?? ""),
          provider: String(row.provider ?? ""),
          prompt_tokens: toInt(row.prompt_tokens),
          output_tokens: toInt(row.output_tokens),
          total_tokens: toInt(row.total_tokens),
          cost_usd: toFloat(row.cost_usd),
          status_code: toInt(row.status_code, 500),
          latency_ms: toInt(row.latency_ms),
          streaming: Boolean(row.streaming),
          error_message: String(row.error_message ?? ""),
          created_at: String(row.created_at ?? ""),
        }))
      : [],
    total: toInt(body.total),
    limit: toInt(body.limit, DEFAULT_LIMIT),
    offset: toInt(body.offset),
  };
}

function getAdminToken(): string {
  const token = process.env.LLMGOPHER_UI_ADMIN_API_KEY?.trim();
  if (!token) {
    throw new Error(AUTH_UNAVAILABLE_MESSAGE);
  }
  return token;
}

async function fetchAnalytics<T>(
  endpoint: string,
  query: URLSearchParams,
  parseBody: (payload: unknown) => T
): Promise<AnalyticsResult<T>> {
  try {
    const token = getAdminToken();
    const response = await fetch(`${endpoint}?${query.toString()}`, {
      cache: "no-store",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.status === 400) {
      return {
        state: "invalid-filter",
        data: parseBody({}),
        message: "The selected filters are invalid. Adjust filters and retry.",
      };
    }
    if (
      response.status === 401 ||
      response.status === 403 ||
      response.status >= 500
    ) {
      return {
        state: "unavailable",
        data: parseBody({}),
        message:
          response.status === 401 || response.status === 403
            ? "Analytics access is unauthorized. Check LLMGOPHER_UI_ADMIN_API_KEY."
            : "Analytics backend is currently unavailable.",
      };
    }
    if (!response.ok) {
      return {
        state: "unavailable",
        data: parseBody({}),
        message: "Analytics request failed.",
      };
    }

    const parsed = parseBody((await response.json()) as unknown);
    const dataRows = (parsed as { data?: unknown[] }).data;
    return {
      state: Array.isArray(dataRows) && dataRows.length === 0 ? "empty" : "ready",
      data: parsed,
    };
  } catch (error) {
    return {
      state: "unavailable",
      data: parseBody({}),
      message: error instanceof Error ? error.message : "Analytics unavailable.",
    };
  }
}

export async function fetchUsageSummary(
  filters: AnalyticsFilters
): Promise<AnalyticsResult<UsageSummaryResponse>> {
  return fetchAnalytics(USAGE_ENDPOINT, buildUsageQuery(filters), parseUsageSummary);
}

export async function fetchDailyUsage(
  filters: AnalyticsFilters
): Promise<AnalyticsResult<DailyUsageResponse>> {
  return fetchAnalytics(
    DAILY_USAGE_ENDPOINT,
    buildDailyUsageQuery(filters),
    parseDailyUsage
  );
}

export async function fetchAuditRecords(
  filters: AnalyticsFilters
): Promise<AnalyticsResult<AuditResponse>> {
  return fetchAnalytics(AUDIT_ENDPOINT, buildAuditQuery(filters), parseAudit);
}

export function formatSmallUSD(value: number): string {
  if (value > 0 && value < 0.01) {
    return "<$0.01";
  }
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  }).format(value);
}
