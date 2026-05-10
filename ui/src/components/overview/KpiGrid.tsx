import { Sparkline } from "./Sparkline";

export interface KpiData {
  requests24h: number;
  spend24h: number;
  errorRate: number;
  avgLatencyMs: number;
  requestsSpark: number[];
  spendSpark: number[];
  errorsSpark: number[];
  latencySpark: number[];
}

function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(2) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return n.toFixed(0);
}

interface KpiCardProps {
  label: string;
  value: string;
  delta?: string;
  hint: string;
  spark: number[];
}

function KpiCard({ label, value, delta, hint, spark }: KpiCardProps) {
  return (
    <div className="kpi">
      <div className="label">{label}</div>
      <div className="value tnum">{value}</div>
      <div className="delta">
        {delta && <b>{delta}</b>}
        <span className="subtle">{hint}</span>
      </div>
      <div className="spark">
        <Sparkline data={spark} width={160} height={40} />
      </div>
    </div>
  );
}

interface KpiGridProps {
  data: KpiData | null;
}

export function KpiGrid({ data }: KpiGridProps) {
  if (!data) {
    return (
      <div className="kpi-grid">
        {["Requests", "Spend", "Error rate", "Avg latency"].map((label) => (
          <div key={label} className="kpi">
            <div className="label">{label}</div>
            <div className="value tnum" style={{ color: "var(--fg-subtle)" }}>
              —
            </div>
            <div className="delta">
              <span className="subtle">no data</span>
            </div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="kpi-grid">
      <KpiCard
        label="Requests"
        value={fmt(data.requests24h)}
        hint="last 24h"
        spark={data.requestsSpark}
      />
      <KpiCard
        label="Spend"
        value={`$${data.spend24h.toFixed(2)}`}
        hint="last 24h"
        spark={data.spendSpark}
      />
      <KpiCard
        label="Error rate"
        value={`${data.errorRate.toFixed(2)}%`}
        hint="lower is better"
        spark={data.errorsSpark}
      />
      <KpiCard
        label="Avg latency"
        value={`${Math.round(data.avgLatencyMs)}ms`}
        hint="across all models"
        spark={data.latencySpark}
      />
    </div>
  );
}
