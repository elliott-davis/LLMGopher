/* global React, LG */
const { useState } = React;
const { Icon, Icons, Pill, StatusPill, Logo, Toggle } = LG;
const D4 = window.LG_DATA;

// ─────────── Logs / Request inspector ───────────
function Logs({ onInspect }) {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Logs</h1>
          <div className="sub">Streaming requests across the gateway. Click any row to inspect prompt, response, and the latency timeline.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.filter}/>Filter</button>
          <button className="btn"><Icon d={Icons.download}/>Export</button>
          <button className="btn primary"><Pill kind="success" icon>live</Pill></button>
        </div>
      </div>

      <div className="card">
        <div className="card-head">
          <div className="row" style={{ gap: 8, flexWrap: "wrap" }}>
            <Pill kind="success" icon>200</Pill>
            <Pill kind="warn">4xx</Pill>
            <Pill kind="danger">5xx</Pill>
            <span style={{ width: 1, height: 18, background: "var(--border)", margin: "0 4px" }}></span>
            <Pill>fallback used</Pill>
            <Pill>guardrail-blocked</Pill>
          </div>
          <div className="right">
            <span className="search" style={{ width: 280 }}><Icon d={Icons.search}/>req_id, key, model…</span>
          </div>
        </div>
        <div className="card-body tight">
          <table className="tbl">
            <thead>
              <tr>
                <th>Time</th><th>Status</th><th>Method</th><th>Route → Model</th>
                <th>Key</th><th className="num">Tokens (in/out)</th><th className="num">Latency</th><th className="num">Cost</th>
              </tr>
            </thead>
            <tbody>
              {D4.requests.map(r => (
                <tr key={r.id} onClick={()=>onInspect(r)} style={{ cursor: "pointer" }}>
                  <td className="id">{r.t}</td>
                  <td>
                    {r.status === 200
                      ? <Pill kind="success">{r.status}</Pill>
                      : r.status >= 500
                        ? <Pill kind="danger">{r.status}</Pill>
                        : <Pill kind="warn">{r.status}</Pill>}
                    {r.fb && <Pill kind="warn" icon>fb</Pill>}
                  </td>
                  <td className="mono" style={{ fontSize: 12, color: "var(--fg-muted)" }}>{r.method}</td>
                  <td>
                    <div className="row" style={{ gap: 8 }}>
                      <Pill>{r.route}</Pill>
                      <span className="subtle">→</span>
                      <Logo kind={LG.getProviderKind(r.model)}/>
                      <span className="mono" style={{ fontSize: 12.5 }}>{r.model}</span>
                    </div>
                  </td>
                  <td className="mono" style={{ fontSize: 12 }}>{r.key}</td>
                  <td className="num tnum subtle">{r.tokens.in.toLocaleString()} / {r.tokens.out.toLocaleString()}</td>
                  <td className="num tnum" style={{ color: r.latency > 2000 ? "var(--warn)" : "inherit" }}>{r.latency}ms</td>
                  <td className="num tnum">${r.cost.toFixed(4)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function RequestInspector({ req, onClose }) {
  if (!req) return null;
  const [tab, setTab] = useState("trace");
  return (
    <div className="drawer-mask" onClick={onClose}>
      <div className="drawer" onClick={e=>e.stopPropagation()}>
        <div className="card-head">
          <div>
            <h3 className="mono">{req.id}</h3>
            <div className="sub mono" style={{ fontSize: 12 }}>{req.t} · {req.method} · {req.key}</div>
          </div>
          <div className="right">
            <button className="btn sm"><Icon d={Icons.copy}/>Copy as cURL</button>
            <button className="btn sm">Replay</button>
            <button className="icon-btn" onClick={onClose}><Icon d={Icons.close}/></button>
          </div>
        </div>
        <div className="tabs">
          {["trace","prompt","response","headers"].map(t=>(
            <button key={t} className={`tab ${tab===t?"active":""}`} onClick={()=>setTab(t)}>{t}</button>
          ))}
        </div>
        <div style={{ flex: 1, overflow: "auto", padding: 22 }}>
          {tab === "trace" && (
            <>
              <div className="kpi-grid" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
                <Stat3b label="Status" value={
                  req.status === 200 ? <Pill kind="success">200 OK</Pill>
                  : req.status >= 500 ? <Pill kind="danger">{req.status}</Pill>
                  : <Pill kind="warn">{req.status}</Pill>
                }/>
                <Stat3b label="Latency" value={`${req.latency}ms`}/>
                <Stat3b label="Tokens" value={`${(req.tokens.in + req.tokens.out).toLocaleString()}`} sub={`${req.tokens.in} in · ${req.tokens.out} out`}/>
                <Stat3b label="Cost" value={`$${req.cost.toFixed(4)}`}/>
              </div>
              <div style={{ height: 18 }}></div>
              <h4 style={hSty2}>Latency timeline</h4>
              <Timeline req={req}/>
              <div style={{ height: 18 }}></div>
              <h4 style={hSty2}>Routing decision</h4>
              <div className="card" style={{ background: "var(--bg-sunken)" }}>
                <div className="card-body" style={{ padding: 14 }}>
                  <div className="row" style={{ gap: 10, fontSize: 12.5 }}>
                    <Pill kind="brand">{req.route}</Pill>
                    <span className="subtle">→</span>
                    {req.fb && <><Pill kind="warn">primary failed</Pill><span className="subtle">→</span></>}
                    <Logo kind={LG.getProviderKind(req.model)}/>
                    <span className="mono">{req.model}</span>
                  </div>
                  {req.fb && <div className="subtle" style={{ fontSize: 12, marginTop: 8 }}>Fallback triggered: primary returned 503 in 2.4s · retried on secondary in 980ms</div>}
                </div>
              </div>
            </>
          )}
          {tab === "prompt" && (
            <div className="code">
{`messages: [
  {
    role: `}<span className="s">"system"</span>{`,
    content: `}<span className="s">"You are a helpful checkout assistant for the Acme storefront. Be concise."</span>{`
  },
  {
    role: `}<span className="s">"user"</span>{`,
    content: `}<span className="s">"My order #84221 hasn't shipped yet. Can you check the status?"</span>{`
  }
],
model: `}<span className="s">"chat-prod"</span>{`,
temperature: `}<span className="n">0.2</span>{`,
max_tokens: `}<span className="n">1024</span>{`,
stream: `}<span className="k">true</span>
            </div>
          )}
          {tab === "response" && (
            <div className="code">
{`{
  id: `}<span className="s">"chatcmpl-9f2a"</span>{`,
  object: `}<span className="s">"chat.completion"</span>{`,
  model: `}<span className="s">"{req.model}"</span>{`,
  choices: [{
    index: `}<span className="n">0</span>{`,
    message: {
      role: `}<span className="s">"assistant"</span>{`,
      content: `}<span className="s">"I see order #84221 — it shipped this morning at 9:14 AM and is currently in transit (UPS tracking 1Z…). Estimated delivery is May 11. Anything else?"</span>{`
    },
    finish_reason: `}<span className="s">"stop"</span>{`
  }],
  usage: {
    prompt_tokens: `}<span className="n">{req.tokens.in}</span>{`,
    completion_tokens: `}<span className="n">{req.tokens.out}</span>{`,
    total_tokens: `}<span className="n">{req.tokens.in + req.tokens.out}</span>{`
  }
}`}
            </div>
          )}
          {tab === "headers" && (
            <div className="code">
{`x-request-id: `}<span className="s">{req.id}</span>{`
x-llmgopher-route: `}<span className="s">{req.route}</span>{`
x-llmgopher-upstream: `}<span className="s">{req.model}</span>{`
x-llmgopher-fallback: `}<span className="s">{req.fb ? "true" : "false"}</span>{`
x-llmgopher-cache: `}<span className="s">"miss"</span>{`
x-llmgopher-cost-usd: `}<span className="s">"{req.cost.toFixed(4)}"</span>{`
x-ratelimit-remaining-rpm: `}<span className="n">412</span>{`
x-ratelimit-remaining-tpm: `}<span className="n">182040</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

const hSty2 = { margin: "0 0 10px", fontSize: 12, fontWeight: 600, color: "var(--fg-muted)", textTransform: "uppercase", letterSpacing: "0.06em" };

function Stat3b({ label, value, sub }) {
  return (
    <div className="kpi" style={{ padding: 14 }}>
      <div className="label">{label}</div>
      <div className="value tnum" style={{ fontSize: 18 }}>{value}</div>
      {sub && <div className="delta"><span className="subtle">{sub}</span></div>}
    </div>
  );
}

// Per-stage waterfall timeline
function Timeline({ req }) {
  const stages = req.fb
    ? [
        { name: "auth",        t: 2,    color: "var(--fg-muted)" },
        { name: "rate-limit",  t: 4,    color: "var(--fg-muted)" },
        { name: "guardrails",  t: 14,   color: "var(--fg-muted)" },
        { name: "primary (timeout)", t: Math.round(req.latency * 0.55), color: "var(--danger)", failed: true },
        { name: "fallback",    t: Math.round(req.latency * 0.42), color: "var(--brand)" },
        { name: "post-process",t: 6,    color: "var(--fg-muted)" },
      ]
    : [
        { name: "auth",        t: 2,                                  color: "var(--fg-muted)" },
        { name: "rate-limit",  t: 4,                                  color: "var(--fg-muted)" },
        { name: "guardrails",  t: 14,                                 color: "var(--fg-muted)" },
        { name: req.model,     t: Math.max(40, req.latency - 26),     color: "var(--brand)" },
        { name: "post-process",t: 6,                                  color: "var(--fg-muted)" },
      ];
  const total = stages.reduce((a,s)=>a+s.t,0);
  let cum = 0;
  return (
    <div style={{ background: "var(--bg-sunken)", border: "1px solid var(--border)", borderRadius: 10, padding: 14 }}>
      {stages.map((s,i) => {
        const startPct = (cum / total) * 100;
        const widthPct = (s.t / total) * 100;
        cum += s.t;
        return (
          <div key={i} style={{ display: "grid", gridTemplateColumns: "150px 1fr 70px", alignItems: "center", gap: 12, padding: "5px 0" }}>
            <span className="mono" style={{ fontSize: 12, color: s.failed ? "var(--danger)" : "var(--fg)" }}>
              {s.failed && "✗ "}{s.name}
            </span>
            <div style={{ position: "relative", height: 14, background: "var(--bg-elev)", border: "1px solid var(--border)", borderRadius: 4 }}>
              <div style={{
                position: "absolute", left: `${startPct}%`, width: `${widthPct}%`, top: 0, bottom: 0,
                background: s.color, opacity: s.failed ? 0.7 : 0.9, borderRadius: 3,
                backgroundImage: s.failed ? "repeating-linear-gradient(45deg, transparent, transparent 4px, rgba(255,255,255,0.2) 4px, rgba(255,255,255,0.2) 8px)" : "none",
              }}></div>
            </div>
            <span className="mono tnum subtle" style={{ fontSize: 11.5, textAlign: "right" }}>{s.t}ms</span>
          </div>
        );
      })}
    </div>
  );
}

// ─────────── Guardrails ───────────
function Guardrails() {
  const [items, setItems] = useState(D4.guardrails);
  const toggle = (id) => setItems(items.map(g => g.id === id ? { ...g, enabled: !g.enabled } : g));
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Guardrails</h1>
          <div className="sub">Pre- and post-processing rules that redact, warn, or block requests. Applied across all routes; per-route overrides supported.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.copy}/>Test playground</button>
          <button className="btn primary"><Icon d={Icons.plus}/>New guardrail</button>
        </div>
      </div>

      <div className="kpi-grid" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
        <Stat3b label="Active guardrails" value={items.filter(g=>g.enabled).length}/>
        <Stat3b label="Hits 24h" value={items.reduce((a,g)=>a+g.hits24h,0).toLocaleString()}/>
        <Stat3b label="Blocked" value="418" sub={<span style={{ color: "var(--danger)" }}>−6% vs prev day</span>}/>
        <Stat3b label="Redacted" value="942" sub="PII tokens removed"/>
      </div>
      <div style={{ height: 18 }}></div>

      <div className="split" style={{ gridTemplateColumns: "1fr 1fr" }}>
        {items.map(g => (
          <div key={g.id} className="card">
            <div className="card-head">
              <div>
                <h3>{g.name}</h3>
                <div className="sub mono" style={{ fontSize: 11 }}>{g.id} · owned by {g.owner}</div>
              </div>
              <div className="right">
                <Toggle on={g.enabled} onChange={()=>toggle(g.id)}/>
              </div>
            </div>
            <div className="card-body" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div className="row" style={{ gap: 14 }}>
                <div>
                  <div className="subtle" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>Action</div>
                  <Pill kind={g.kind === "block" ? "danger" : g.kind === "warn" ? "warn" : "brand"}>{g.kind}</Pill>
                </div>
                <div>
                  <div className="subtle" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>Scope</div>
                  <Pill>{g.scope}</Pill>
                </div>
                <div>
                  <div className="subtle" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>Hits 24h</div>
                  <div className="mono tnum" style={{ fontSize: 16, fontWeight: 600 }}>{g.hits24h.toLocaleString()}</div>
                </div>
              </div>
              <button className="btn sm">Configure</button>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

// ─────────── Audit log ───────────
function Audit() {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Audit log</h1>
          <div className="sub">Every administrative action — key creation, route changes, budget edits — is appended here, immutable.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.filter}/>Filter</button>
          <button className="btn"><Icon d={Icons.download}/>Export JSON</button>
        </div>
      </div>
      <div className="card">
        <div className="card-head"><h3>Last 24h</h3><div className="right"><Pill icon kind="success">tamper-evident</Pill></div></div>
        <div className="card-body tight">
          <table className="tbl">
            <thead><tr><th>Time</th><th>Actor</th><th>Action</th><th>Target</th><th>IP</th></tr></thead>
            <tbody>
              {D4.audit.map((a,i)=>(
                <tr key={i}>
                  <td className="id">{a.t}</td>
                  <td>
                    <div className="row" style={{ gap: 8 }}>
                      {a.who === "system" ? <Pill>system</Pill> : (
                        <>
                          <span style={{ width: 22, height: 22, borderRadius: "50%", background: "linear-gradient(135deg, oklch(0.7 0.1 200), oklch(0.65 0.12 280))", color: "white", display: "grid", placeItems: "center", fontSize: 10, fontWeight: 700 }}>
                            {a.who[0].toUpperCase()}
                          </span>
                          <span className="mono" style={{ fontSize: 12.5 }}>{a.who}</span>
                        </>
                      )}
                    </div>
                  </td>
                  <td><Pill kind={a.action.includes("trip") || a.action.includes("degrade") ? "warn" : "brand"}>{a.action}</Pill></td>
                  <td className="mono" style={{ fontSize: 12.5 }}>{a.target}</td>
                  <td className="mono subtle" style={{ fontSize: 12 }}>{a.ip}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

// ─────────── Settings ───────────
function Settings() {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Settings</h1>
          <div className="sub">Org-wide configuration: auth, secret backend, regions, integrations.</div>
        </div>
      </div>

      <div className="split split-2">
        <div className="card">
          <div className="card-head"><h3>Authentication</h3><span className="sub">SSO + access</span></div>
          <div className="card-body">
            <SettingRow title="SSO provider" value={<Pill icon kind="success">Okta · llmgopher.okta.com</Pill>} action="Reconfigure"/>
            <SettingRow title="Default role for new users" value={<Pill>viewer</Pill>} action="Edit"/>
            <SettingRow title="Require MFA for admins" value={<Toggle on={true} onChange={()=>{}}/>}/>
            <SettingRow title="Session lifetime" value={<span className="mono tnum">8h</span>} action="Edit"/>
          </div>
        </div>
        <div className="card">
          <div className="card-head"><h3>Secrets</h3><span className="sub">where API keys are stored</span></div>
          <div className="card-body">
            <SettingRow title="Backend" value={<Pill icon kind="success">HashiCorp Vault · vault.internal</Pill>} action="Edit"/>
            <SettingRow title="Encryption key" value={<span className="mono subtle">kms://aws/llmgopher-prod</span>}/>
            <SettingRow title="Rotate provider creds" value={<span className="mono">every 90 days</span>} action="Edit"/>
          </div>
        </div>
        <div className="card">
          <div className="card-head"><h3>Storage</h3><span className="sub">audit, budgets, rate-limit</span></div>
          <div className="card-body">
            <SettingRow title="Postgres (audit + budgets)" value={<span className="mono subtle">postgres://prod-pg-east/llmgopher</span>} action="Test"/>
            <SettingRow title="Redis (rate-limit + cache)" value={<span className="mono subtle">redis://prod-cache:6379</span>} action="Test"/>
            <SettingRow title="Object store (request bodies)" value={<span className="mono subtle">s3://llmgopher-logs</span>} action="Edit"/>
          </div>
        </div>
        <div className="card">
          <div className="card-head"><h3>Integrations</h3></div>
          <div className="card-body">
            <SettingRow title="Slack alerts" value={<Pill icon kind="success">#llm-platform</Pill>} action="Edit"/>
            <SettingRow title="PagerDuty" value={<Pill>not connected</Pill>} action="Connect"/>
            <SettingRow title="Datadog metrics" value={<Pill icon kind="success">connected</Pill>} action="Edit"/>
            <SettingRow title="OpenTelemetry export" value={<Pill icon kind="success">otlp://otel-collector:4317</Pill>} action="Edit"/>
          </div>
        </div>
      </div>
    </>
  );
}

function SettingRow({ title, value, action }) {
  return (
    <div className="row between" style={{ padding: "10px 0", borderBottom: "1px solid var(--border)" }}>
      <div>
        <div style={{ fontWeight: 500, fontSize: 13.5 }}>{title}</div>
      </div>
      <div className="row" style={{ gap: 12 }}>
        {value}
        {action && <button className="btn sm">{action}</button>}
      </div>
    </div>
  );
}

window.LG.Logs = Logs;
window.LG.RequestInspector = RequestInspector;
window.LG.Guardrails = Guardrails;
window.LG.Audit = Audit;
window.LG.Settings = Settings;
