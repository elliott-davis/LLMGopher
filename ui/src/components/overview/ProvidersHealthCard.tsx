import Link from "next/link";
import { Provider } from "@/lib/types";

function ProviderLogo({ name }: { name: string }) {
  const lower = name.toLowerCase();
  let kind = "openai";
  let letter = "O";
  if (lower.includes("anthropic")) { kind = "anthropic"; letter = "A"; }
  else if (lower.includes("vertex") || lower.includes("google") || lower.includes("gemini")) { kind = "google"; letter = "G"; }
  else if (lower.includes("mistral")) { kind = "mistral"; letter = "M"; }
  else if (lower.includes("bedrock")) { kind = "bedrock"; letter = "B"; }
  else if (lower.includes("vllm") || lower.includes("vl")) { kind = "vllm"; letter = "v"; }
  return <span className={`logo ${kind}`}>{letter}</span>;
}

interface ProvidersHealthCardProps {
  providers: Provider[];
}

export function ProvidersHealthCard({ providers }: ProvidersHealthCardProps) {
  return (
    <div className="card">
      <div className="card-head">
        <h3>Providers</h3>
        <span className="sub">{providers.length} configured</span>
        <div className="right">
          <Link href="/providers" className="btn ghost sm">
            Manage
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </Link>
        </div>
      </div>
      <div className="card-body" style={{ display: "flex", flexDirection: "column", gap: 0 }}>
        {providers.length === 0 ? (
          <p style={{ color: "var(--fg-muted)", fontSize: 13, margin: 0 }}>
            No providers configured.
          </p>
        ) : (
          providers.map((p, i) => (
            <div
              key={p.id}
              className="row between"
              style={{
                padding: "10px 4px",
                borderBottom: i < providers.length - 1 ? "1px solid var(--border)" : "none",
              }}
            >
              <div className="row" style={{ gap: 10 }}>
                <ProviderLogo name={p.name} />
                <div>
                  <div style={{ fontWeight: 500, fontSize: 13 }}>{p.name}</div>
                  <div
                    className="subtle mono"
                    style={{ fontSize: 11, marginTop: 1 }}
                  >
                    {p.auth_type}
                  </div>
                </div>
              </div>
              <span
                className="pill"
                style={{ fontSize: 11 }}
                title={p.has_credentials ? "Credentials configured" : "No credentials"}
              >
                {p.has_credentials ? "configured" : "no credentials"}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
