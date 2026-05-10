import Link from "next/link";

export interface ModelUsageRow {
  group: string;        // model name/alias
  requests: number;
  cost_usd: number;
  errors: number;
  avg_latency_ms: number;
}

function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return n.toFixed(0);
}

interface TopModelsCardProps {
  rows: ModelUsageRow[];
}

export function TopModelsCard({ rows }: TopModelsCardProps) {
  const sorted = [...rows].sort((a, b) => b.requests - a.requests).slice(0, 6);

  return (
    <div className="card">
      <div className="card-head">
        <h3>Top models</h3>
        <span className="sub">by request volume · 24h</span>
        <div className="right">
          <Link href="/routes" className="btn ghost sm">
            View all
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </Link>
        </div>
      </div>
      <div className="card-body tight">
        {sorted.length === 0 ? (
          <div style={{ padding: 18, color: "var(--fg-muted)", fontSize: 13 }}>
            No traffic data for the last 24h.
          </div>
        ) : (
          <table className="tbl">
            <thead>
              <tr>
                <th>Model</th>
                <th className="num">Requests</th>
                <th className="num">Avg lat</th>
                <th className="num">Cost</th>
                <th className="num">Errors</th>
              </tr>
            </thead>
            <tbody>
              {sorted.map((r) => {
                const errRate =
                  r.requests > 0
                    ? ((r.errors / r.requests) * 100).toFixed(1)
                    : "0.0";
                const isHighErr = r.requests > 0 && r.errors / r.requests > 0.01;
                return (
                  <tr key={r.group}>
                    <td>
                      <span className="mono" style={{ fontWeight: 500, fontSize: 12.5 }}>
                        {r.group}
                      </span>
                    </td>
                    <td className="num tnum">{fmt(r.requests)}</td>
                    <td className="num tnum">
                      {Math.round(r.avg_latency_ms)}ms
                    </td>
                    <td className="num tnum">${r.cost_usd.toFixed(4)}</td>
                    <td
                      className="num tnum"
                      style={{ color: isHighErr ? "var(--warn)" : "inherit" }}
                    >
                      {errRate}%
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
