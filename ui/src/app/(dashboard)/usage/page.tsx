import AnalyticsState from "@/components/usage/AnalyticsState";
import AuditLogTable from "@/components/usage/AuditLogTable";
import UsageFilterForm from "@/components/usage/UsageFilterForm";
import UsageSummaryTable from "@/components/usage/UsageSummaryTable";
import UsageTrendTable from "@/components/usage/UsageTrendTable";
import {
  fetchAuditRecords,
  fetchDailyUsage,
  fetchUsageSummary,
  parseAnalyticsFilters,
} from "@/lib/analytics";
import { AnalyticsFilters } from "@/lib/types";

type SearchParamsRecord = Record<string, string | string[] | undefined>;

interface UsagePageProps {
  searchParams?: Promise<SearchParamsRecord>;
}

function displayFilters(source: SearchParamsRecord): AnalyticsFilters {
  const groupBy = String(source.group_by ?? "").trim();
  const status = String(source.status ?? "").trim();
  return {
    from: String(source.from ?? "").trim() || undefined,
    to: String(source.to ?? "").trim() || undefined,
    group_by:
      groupBy === "provider" || groupBy === "api_key" || groupBy === "model"
        ? groupBy
        : "model",
    api_key_id: String(source.api_key_id ?? "").trim() || undefined,
    model: String(source.model ?? "").trim() || undefined,
    provider: String(source.provider ?? "").trim() || undefined,
    status: status === "success" || status === "error" ? status : undefined,
    limit: Number(source.limit ?? 25) > 0 ? Number(source.limit) : 25,
    offset: Number(source.offset ?? 0) >= 0 ? Number(source.offset) : 0,
  };
}

export default async function UsagePage({ searchParams }: UsagePageProps) {
  const resolvedSearchParams = (await searchParams) ?? {};

  let filters: AnalyticsFilters;
  let filterError: string | null = null;
  try {
    filters = parseAnalyticsFilters(resolvedSearchParams);
  } catch (error) {
    filters = displayFilters(resolvedSearchParams);
    filterError = error instanceof Error ? error.message : "Invalid filters";
  }

  const [usage, daily, audit] = filterError
    ? [null, null, null]
    : await Promise.all([
        fetchUsageSummary(filters),
        fetchDailyUsage(filters),
        fetchAuditRecords(filters),
      ]);

  return (
    <div className="mx-auto w-full max-w-7xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Usage & Audit</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Review usage summaries, daily trends, and audit records with shared URL
          filters.
        </p>
      </div>

      <UsageFilterForm filters={filters} />

      {filterError ? (
        <>
          <AnalyticsState
            title="Grouped Usage Summary"
            state="invalid-filter"
            message={filterError}
          />
          <AnalyticsState
            title="Daily Usage Trend"
            state="invalid-filter"
            message={filterError}
          />
          <AnalyticsState
            title="Audit Log Search"
            state="invalid-filter"
            message={filterError}
          />
        </>
      ) : (
        <>
          {usage?.state === "ready" ? (
            <UsageSummaryTable rows={usage.data.data} />
          ) : (
            <AnalyticsState
              title="Grouped Usage Summary"
              state={usage?.state ?? "unavailable"}
              message={usage?.message}
            />
          )}

          {daily?.state === "ready" ? (
            <UsageTrendTable rows={daily.data.data} />
          ) : (
            <AnalyticsState
              title="Daily Usage Trend"
              state={daily?.state ?? "unavailable"}
              message={daily?.message}
            />
          )}

          {audit?.state === "ready" ? (
            <AuditLogTable
              rows={audit.data.data}
              total={audit.data.total}
              limit={audit.data.limit}
              offset={audit.data.offset}
              filters={filters}
            />
          ) : (
            <AnalyticsState
              title="Audit Log Search"
              state={audit?.state ?? "unavailable"}
              message={audit?.message}
            />
          )}
        </>
      )}
    </div>
  );
}
