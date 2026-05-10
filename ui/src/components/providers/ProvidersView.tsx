"use client";

import { useState } from "react";
import { ProviderLogo } from "./ProviderLogo";
import { AddProviderDrawer } from "./AddProviderDrawer";
import ProviderRowActions from "@/components/ProviderRowActions";
import type { Provider } from "@/lib/types";

interface ProvidersViewProps {
  providers: Provider[];
  modelCount: number;
}

function PlusIcon() {
  return (
    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M12 5v14M5 12h14" />
    </svg>
  );
}

export function ProvidersView({ providers, modelCount }: ProvidersViewProps) {
  const [drawerOpen, setDrawerOpen] = useState(false);

  const withCreds = providers.filter((p) => p.has_credentials).length;
  const withoutCreds = providers.length - withCreds;

  const kpis = [
    { label: "Providers", value: providers.length, sub: "all environments" },
    { label: "With credentials", value: withCreds, sub: "ready to route" },
    { label: "Missing credentials", value: withoutCreds, sub: "needs attention" },
    { label: "Models exposed", value: modelCount, sub: "across all providers" },
  ];

  function formatDate(iso?: string) {
    if (!iso) return "—";
    try {
      return new Date(iso).toLocaleDateString("en-US", {
        month: "short", day: "numeric", year: "numeric",
      });
    } catch { return "—"; }
  }

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Providers</h1>
          <div className="sub">
            Upstream LLM endpoints. Each provider can host multiple models.
          </div>
        </div>
        <div className="actions">
          <button className="btn primary" data-testid="add-provider" onClick={() => setDrawerOpen(true)}>
            <PlusIcon />
            Add provider
          </button>
        </div>
      </div>

      {/* KPI mini-strip */}
      <div
        className="kpi-grid"
        style={{ gridTemplateColumns: "repeat(4, 1fr)", marginBottom: 18 }}
      >
        {kpis.map((k) => (
          <div key={k.label} className="kpi" style={{ padding: "12px 16px" }}>
            <div className="label">{k.label}</div>
            <div className="value tnum" style={{ fontSize: 22 }}>{k.value}</div>
            <div className="delta">
              <span className="subtle">{k.sub}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Table */}
      <div className="card">
        <div className="card-head">
          <h3>All providers</h3>
          <span className="sub">{providers.length} configured</span>
        </div>
        <div className="card-body tight">
          <table className="tbl">
            <thead>
              <tr>
                <th>Provider</th>
                <th>Base URL</th>
                <th>Auth type</th>
                <th>Credentials</th>
                <th>Updated</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {providers.length === 0 ? (
                <tr>
                  <td
                    colSpan={6}
                    style={{ color: "var(--fg-muted)", textAlign: "center", padding: 32 }}
                  >
                    No providers configured. Click &ldquo;Add provider&rdquo; to get started.
                  </td>
                </tr>
              ) : (
                providers.map((p) => (
                  <tr key={p.id} data-testid={`provider-row-${p.id}`}>
                    <td>
                      <div className="row" style={{ gap: 10 }}>
                        <ProviderLogo name={p.name} />
                        <div>
                          <div style={{ fontWeight: 500 }}>{p.name}</div>
                          <div className="id" style={{ fontSize: 11 }}>{p.id}</div>
                        </div>
                      </div>
                    </td>
                    <td
                      className="mono"
                      style={{ fontSize: 12, color: "var(--fg-muted)", maxWidth: 260, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
                    >
                      {p.base_url}
                    </td>
                    <td>
                      <span className="pill">{p.auth_type}</span>
                    </td>
                    <td>
                      <span
                        className={`pill ${p.has_credentials ? "success" : "warn"}`}
                      >
                        <span className="dot" />
                        {p.has_credentials ? "configured" : "missing"}
                      </span>
                    </td>
                    <td style={{ color: "var(--fg-muted)", fontSize: 12 }}>
                      {formatDate(p.updated_at)}
                    </td>
                    <td>
                      <ProviderRowActions provider={p} />
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      <AddProviderDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      />
    </>
  );
}
