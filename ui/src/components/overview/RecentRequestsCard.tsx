import { AuditRecord } from "@/lib/types";

function StatusPill({ code }: { code: number }) {
  const kind =
    code === 200 ? "success" : code >= 500 ? "danger" : code >= 400 ? "warn" : "";
  return <span className={`pill ${kind}`}>{code}</span>;
}

function ModelLogo({ model, provider }: { model: string; provider: string }) {
  const src = (model + provider).toLowerCase();
  let kind = "openai";
  let letter = "O";
  if (src.includes("anthropic") || src.includes("claude")) {
    kind = "anthropic"; letter = "A";
  } else if (src.includes("vertex") || src.includes("google") || src.includes("gemini")) {
    kind = "google"; letter = "G";
  } else if (src.includes("mistral")) {
    kind = "mistral"; letter = "M";
  } else if (src.includes("bedrock") || src.includes("llama")) {
    kind = "bedrock"; letter = "B";
  } else if (src.includes("vllm")) {
    kind = "vllm"; letter = "v";
  }
  return <span className={`logo ${kind}`}>{letter}</span>;
}

function formatTime(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleTimeString("en-US", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } catch {
    return iso;
  }
}

interface RecentRequestsCardProps {
  records: AuditRecord[];
}

export function RecentRequestsCard({ records }: RecentRequestsCardProps) {
  return (
    <div className="card">
      <div className="card-head">
        <h3>Recent requests</h3>
        <span className="sub">from audit log</span>
        <div className="right">
          <span className="pill success" style={{ gap: 6 }}>
            <span
              style={{
                width: 6,
                height: 6,
                borderRadius: "50%",
                background: "var(--success)",
                boxShadow:
                  "0 0 0 3px color-mix(in oklch, var(--success) 25%, transparent)",
              }}
            />
            live
          </span>
        </div>
      </div>
      <div
        className="card-body tight"
        style={{ maxHeight: 320, overflow: "auto" }}
      >
        {records.length === 0 ? (
          <div style={{ padding: 18, color: "var(--fg-muted)", fontSize: 13 }}>
            No recent requests.
          </div>
        ) : (
          <table className="tbl">
            <tbody>
              {records.map((r) => (
                <tr key={r.id}>
                  <td
                    className="id"
                    style={{ width: 88, fontSize: 11, whiteSpace: "nowrap" }}
                  >
                    {formatTime(r.created_at)}
                  </td>
                  <td style={{ width: 28 }}>
                    <ModelLogo model={r.model} provider={r.provider} />
                  </td>
                  <td>
                    <span className="mono" style={{ fontSize: 12 }}>
                      {r.model}
                    </span>
                  </td>
                  <td>
                    <StatusPill code={r.status_code} />
                  </td>
                  <td
                    className="num tnum subtle"
                    style={{ fontSize: 12, whiteSpace: "nowrap" }}
                  >
                    {r.latency_ms}ms
                  </td>
                  <td
                    className="num tnum subtle"
                    style={{ fontSize: 12 }}
                  >
                    ${r.cost_usd.toFixed(4)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
