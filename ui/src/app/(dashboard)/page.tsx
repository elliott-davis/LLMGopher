import { KpiGrid, type KpiData } from "@/components/overview/KpiGrid";
import { TopModelsCard, type ModelUsageRow } from "@/components/overview/TopModelsCard";
import { ProvidersHealthCard } from "@/components/overview/ProvidersHealthCard";
import { RecentRequestsCard } from "@/components/overview/RecentRequestsCard";
import {
  Provider,
  UsageSummaryResponse,
  DailyUsageResponse,
  AuditResponse,
} from "@/lib/types";

const GATEWAY = process.env.LLMGOPHER_GATEWAY_BASE ?? "http://gateway:8080";

function adminToken(): string {
  return process.env.LLMGOPHER_UI_ADMIN_API_KEY?.trim() ?? "";
}

async function fetchJSON<T>(
  url: string,
  token?: string
): Promise<T | null> {
  try {
    const res = await fetch(url, {
      cache: "no-store",
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    if (!res.ok) return null;
    return res.json() as Promise<T>;
  } catch {
    return null;
  }
}

function buildKpis(
  usage: UsageSummaryResponse | null,
  daily: DailyUsageResponse | null
): KpiData | null {
  if (!usage || !usage.data || usage.data.length === 0) return null;

  const totals = usage.data.reduce(
    (acc, row) => ({
      requests: acc.requests + row.requests,
      cost: acc.cost + row.cost_usd,
      errors: acc.errors + row.errors,
      latencySum: acc.latencySum + row.avg_latency_ms * row.requests,
    }),
    { requests: 0, cost: 0, errors: 0, latencySum: 0 }
  );

  const errorRate =
    totals.requests > 0 ? (totals.errors / totals.requests) * 100 : 0;
  const avgLatency =
    totals.requests > 0 ? totals.latencySum / totals.requests : 0;

  // Build sparklines from daily data if available (7-day window)
  const dailyPts = daily?.data ?? [];
  const requestsSpark = dailyPts.map((d) => d.requests);
  const spendSpark = dailyPts.map((d) => d.cost_usd);

  return {
    requests24h: totals.requests,
    spend24h: totals.cost,
    errorRate,
    avgLatencyMs: avgLatency,
    requestsSpark,
    spendSpark,
    errorsSpark: [],
    latencySpark: [],
  };
}

function buildModelRows(usage: UsageSummaryResponse | null): ModelUsageRow[] {
  if (!usage?.data) return [];
  return usage.data.map((row) => ({
    group: row.group,
    requests: row.requests,
    cost_usd: row.cost_usd,
    errors: row.errors,
    avg_latency_ms: row.avg_latency_ms,
  }));
}

export default async function OverviewPage() {
  const now = new Date();
  const from24h = new Date(now.getTime() - 24 * 60 * 60 * 1000);
  const from7d = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
  const token = adminToken();

  const [providers, usage24h, daily7d, audit] = await Promise.all([
    fetchJSON<Provider[]>(`${GATEWAY}/v1/admin/providers`),
    fetchJSON<UsageSummaryResponse>(
      `${GATEWAY}/v1/admin/usage?group_by=model&from=${from24h.toISOString()}&to=${now.toISOString()}`,
      token
    ),
    fetchJSON<DailyUsageResponse>(
      `${GATEWAY}/v1/admin/usage/daily?from=${from7d.toISOString()}&to=${now.toISOString()}`,
      token
    ),
    fetchJSON<AuditResponse>(
      `${GATEWAY}/v1/admin/audit?limit=10`,
      token
    ),
  ]);

  const kpis = buildKpis(usage24h, daily7d);
  const modelRows = buildModelRows(usage24h);
  const recentRecords = audit?.data ?? [];

  return (
    <div>
      <div className="page-head">
        <div>
          <h1>Overview</h1>
          <div className="sub">
            Real-time traffic across all routes · last 24h
          </div>
        </div>
        <div className="actions">
          <select className="btn" style={{ paddingRight: 8, cursor: "pointer" }}>
            <option>Last 24h</option>
            <option>Last 7d</option>
            <option>Last 30d</option>
          </select>
          <button className="btn primary">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
            </svg>
            Export
          </button>
        </div>
      </div>

      <KpiGrid data={kpis} />

      <div style={{ height: 22 }} />

      <div className="split split-2">
        <TopModelsCard rows={modelRows} />
        <ProvidersHealthCard providers={providers ?? []} />
      </div>

      <div style={{ height: 16 }} />

      <div className="split split-2">
        <RecentRequestsCard records={recentRecords} />
        <BudgetGlanceCard
          keyCount={(providers ?? []).length}
          spend={kpis?.spend24h ?? null}
        />
      </div>
    </div>
  );
}

// Inline budget glance — simplified until /admin/teams is available
function BudgetGlanceCard({
  keyCount,
  spend,
}: {
  keyCount: number;
  spend: number | null;
}) {
  return (
    <div className="card">
      <div className="card-head">
        <h3>Spend (24h)</h3>
        <span className="sub">all models</span>
      </div>
      <div className="card-body">
        {spend !== null ? (
          <>
            <div
              className="value mono tnum"
              style={{
                fontSize: 28,
                fontWeight: 600,
                letterSpacing: "-0.02em",
              }}
            >
              ${spend.toFixed(4)}
            </div>
            <div className="subtle" style={{ fontSize: 12, marginTop: 4 }}>
              aggregated across all models
            </div>
          </>
        ) : (
          <>
            <div
              className="value mono tnum"
              style={{
                fontSize: 28,
                fontWeight: 600,
                letterSpacing: "-0.02em",
                color: "var(--fg-subtle)",
              }}
            >
              —
            </div>
            <div className="subtle" style={{ fontSize: 12, marginTop: 4 }}>
              set LLMGOPHER_UI_ADMIN_API_KEY to enable spend tracking
            </div>
          </>
        )}
        <div
          style={{
            marginTop: 16,
            paddingTop: 14,
            borderTop: "1px solid var(--border)",
            display: "flex",
            gap: 20,
          }}
        >
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-subtle)", textTransform: "uppercase", letterSpacing: "0.06em" }}>Providers</div>
            <div className="tnum" style={{ fontWeight: 600, fontSize: 18, marginTop: 2 }}>{keyCount}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
