/* global React, LG */
const { useState, useMemo, useEffect } = React;
const { Icon, Icons, Spark, Pill, StatusPill, Logo, Toggle } = LG;
const D2 = window.LG_DATA;

// ─────────── Providers ───────────
function Providers({ onAdd }) {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Providers</h1>
          <div className="sub">Upstream LLM endpoints. Each provider can host multiple deployments and models.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.filter}/>Filter</button>
          <button className="btn primary" onClick={onAdd}><Icon d={Icons.plus}/>Add provider</button>
        </div>
      </div>

      <div className="kpi-grid" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
        {[
          { l: "Providers", v: D2.providers.length, s: "all environments" },
          { l: "Healthy", v: D2.providers.filter(p=>p.status==="healthy").length, s: "passing health checks" },
          { l: "Degraded / offline", v: D2.providers.filter(p=>p.status!=="healthy").length, s: "needs attention" },
          { l: "Models exposed", v: D2.models.length, s: "across all providers" },
        ].map((k,i)=>(
          <div key={i} className="kpi" style={{ padding: "12px 16px" }}>
            <div className="label">{k.l}</div>
            <div className="value tnum" style={{ fontSize: 22 }}>{k.v}</div>
            <div className="delta"><span className="subtle">{k.s}</span></div>
          </div>
        ))}
      </div>

      <div style={{ height: 18 }}></div>

      <div className="card">
        <div className="card-head">
          <h3>All providers</h3>
          <div className="right">
            <span className="search" style={{ width: 220 }}><Icon d={Icons.search}/>filter…</span>
          </div>
        </div>
        <div className="card-body tight">
          <table className="tbl">
            <thead>
              <tr>
                <th>Provider</th><th>Base URL</th><th>Region</th>
                <th className="num">Deployments</th><th className="num">p50 / p95</th>
                <th className="num">Errors 24h</th><th>Status</th><th></th>
              </tr>
            </thead>
            <tbody>
              {D2.providers.map(p=>(
                <tr key={p.id}>
                  <td>
                    <div className="row" style={{ gap: 10 }}>
                      <Logo kind={p.logo}/>
                      <div>
                        <div style={{ fontWeight: 500 }}>{p.name}</div>
                        <div className="id" style={{ fontSize: 11 }}>{p.id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="mono" style={{ fontSize: 12, color: "var(--fg-muted)" }}>{p.base}</td>
                  <td><Pill>{p.region}</Pill></td>
                  <td className="num tnum">{p.deployments}</td>
                  <td className="num tnum">{p.latencyP50 || "—"} / {p.latencyP95 || "—"}</td>
                  <td className="num tnum" style={{ color: p.errors24h > 20 ? "var(--warn)" : "inherit" }}>{p.errors24h}</td>
                  <td><StatusPill status={p.status}/></td>
                  <td><button className="icon-btn" style={{ width: 26, height: 26 }}><Icon d={Icons.more}/></button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

// Add Provider drawer — multi-step
function AddProviderDrawer({ open, onClose }) {
  const [step, setStep] = useState(0);
  const [kind, setKind] = useState("openai");
  if (!open) return null;
  const kinds = [
    { id: "openai", label: "OpenAI", logo: "openai", desc: "OpenAI API · gpt-4o family" },
    { id: "anthropic", label: "Anthropic", logo: "anthropic", desc: "Anthropic API · claude family" },
    { id: "google", label: "Vertex AI", logo: "google", desc: "Google Vertex · gemini family" },
    { id: "bedrock", label: "AWS Bedrock", logo: "bedrock", desc: "Bedrock · multi-tenant" },
    { id: "mistral", label: "Mistral", logo: "mistral", desc: "Mistral API · EU residency" },
    { id: "vllm", label: "vLLM (self-hosted)", logo: "vllm", desc: "OpenAI-compatible endpoint" },
  ];
  return (
    <div className="drawer-mask" onClick={onClose}>
      <div className="drawer" onClick={e=>e.stopPropagation()}>
        <div className="card-head" style={{ borderBottom: "1px solid var(--border)" }}>
          <h3>Add provider</h3>
          <div className="sub">Step {step+1} of 3</div>
          <div className="right">
            <button className="icon-btn" onClick={onClose}><Icon d={Icons.close}/></button>
          </div>
        </div>
        <div style={{ flex: 1, overflow: "auto", padding: 22 }}>
          {step === 0 && (
            <>
              <h4 style={{ margin: "0 0 12px", fontSize: 13, fontWeight: 600 }}>Choose a provider type</h4>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
                {kinds.map(k=>(
                  <div key={k.id} onClick={()=>setKind(k.id)} className="card" style={{
                    padding: 14, cursor: "pointer",
                    borderColor: kind===k.id ? "var(--brand)" : "var(--border)",
                    boxShadow: kind===k.id ? "0 0 0 3px color-mix(in oklch, var(--brand) 20%, transparent)" : "var(--shadow-1)"
                  }}>
                    <div className="row" style={{ gap: 10 }}>
                      <Logo kind={k.logo} size="lg"/>
                      <div>
                        <div style={{ fontWeight: 600 }}>{k.label}</div>
                        <div className="subtle" style={{ fontSize: 12 }}>{k.desc}</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
          {step === 1 && (
            <>
              <h4 style={{ margin: "0 0 12px", fontSize: 13, fontWeight: 600 }}>Credentials & endpoint</h4>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
                <div className="field" style={{ gridColumn: "1 / -1" }}><label>Display name</label><input defaultValue="OpenAI · prod"/></div>
                <div className="field" style={{ gridColumn: "1 / -1" }}><label>Base URL</label><input className="mono" defaultValue="https://api.openai.com/v1"/></div>
                <div className="field" style={{ gridColumn: "1 / -1" }}><label>API key (stored encrypted in vault)</label><input type="password" placeholder="sk-…"/></div>
                <div className="field"><label>Region</label><select><option>us-east</option><option>us-west</option><option>eu-west</option></select></div>
                <div className="field"><label>Timeout (ms)</label><input className="mono tnum" defaultValue="30000"/></div>
              </div>
              <div style={{ height: 18 }}></div>
              <div className="card" style={{ background: "var(--bg-sunken)", borderStyle: "dashed" }}>
                <div className="card-body">
                  <div className="row" style={{ gap: 10 }}>
                    <Pill kind="success" icon>connection ok</Pill>
                    <span className="subtle" style={{ fontSize: 12 }}>tested 312ms ago · GET /v1/models returned 47 models</span>
                  </div>
                </div>
              </div>
            </>
          )}
          {step === 2 && (
            <>
              <h4 style={{ margin: "0 0 12px", fontSize: 13, fontWeight: 600 }}>Pick models to expose</h4>
              <div className="card" style={{ background: "var(--bg-sunken)" }}>
                <div className="card-body tight">
                  <table className="tbl">
                    <thead><tr><th></th><th>Upstream model</th><th>Context</th><th className="num">Input $/M</th><th className="num">Output $/M</th></tr></thead>
                    <tbody>
                      {[
                        ["gpt-4o", 128000, 2.5, 10.0],
                        ["gpt-4o-mini", 128000, 0.15, 0.6],
                        ["gpt-4-turbo", 128000, 10.0, 30.0],
                        ["o1", 200000, 15.0, 60.0],
                        ["o1-mini", 128000, 3.0, 12.0],
                        ["text-embedding-3-large", 8192, 0.13, 0],
                      ].map(([m, c, i, o], idx) => (
                        <tr key={m}>
                          <td><input type="checkbox" defaultChecked={idx < 2}/></td>
                          <td className="mono">{m}</td>
                          <td className="mono tnum">{(c/1000)|0}k</td>
                          <td className="num tnum">${i}</td>
                          <td className="num tnum">${o}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </>
          )}
        </div>
        <div className="row between" style={{ padding: 14, borderTop: "1px solid var(--border)" }}>
          <button className="btn ghost" onClick={onClose}>Cancel</button>
          <div className="row" style={{ gap: 8 }}>
            {step > 0 && <button className="btn" onClick={()=>setStep(s=>s-1)}>Back</button>}
            {step < 2 && <button className="btn primary" onClick={()=>setStep(s=>s+1)}>Continue</button>}
            {step === 2 && <button className="btn primary" onClick={onClose}><Icon d={Icons.zap}/>Create provider</button>}
          </div>
        </div>
      </div>
    </div>
  );
}

// ─────────── Routes — with visualizer ───────────
function Routes() {
  const [selected, setSelected] = useState(D2.routes[0].id);
  const route = D2.routes.find(r => r.id === selected);
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Routes</h1>
          <div className="sub">Logical model names that map to one or more upstream deployments via fallback, weighted, or latency-based strategies.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.copy}/>Import YAML</button>
          <button className="btn primary"><Icon d={Icons.plus}/>New route</button>
        </div>
      </div>

      <div className="split" style={{ gridTemplateColumns: "320px 1fr" }}>
        <div className="card">
          <div className="card-head"><h3>All routes</h3><span className="sub">{D2.routes.length}</span></div>
          <div style={{ display: "flex", flexDirection: "column" }}>
            {D2.routes.map(r => (
              <div key={r.id} onClick={()=>setSelected(r.id)} style={{
                padding: "12px 16px", cursor: "pointer",
                borderBottom: "1px solid var(--border)",
                background: selected === r.id ? "var(--bg-hover)" : "transparent",
                borderLeft: selected === r.id ? "2px solid var(--brand)" : "2px solid transparent",
              }}>
                <div className="row between">
                  <span className="mono" style={{ fontWeight: 600 }}>{r.name}</span>
                  <Pill>{r.strategy}</Pill>
                </div>
                <div className="subtle" style={{ fontSize: 11.5, marginTop: 4 }}>
                  {r.members.length} member{r.members.length>1?"s":""} · {r.rpm.toLocaleString()} rpm · {r.errRate}% err
                </div>
              </div>
            ))}
          </div>
        </div>

        <RouteDetail route={route}/>
      </div>
    </>
  );
}

function RouteDetail({ route }) {
  return (
    <div className="card">
      <div className="card-head">
        <div>
          <h3 className="mono">{route.name}</h3>
          <div className="sub">{route.desc}</div>
        </div>
        <div className="right">
          <Pill kind="success" icon>active</Pill>
          <button className="btn sm"><Icon d={Icons.copy}/>YAML</button>
          <button className="btn sm primary">Edit</button>
        </div>
      </div>
      <div className="card-body">
        <div className="row" style={{ gap: 14, marginBottom: 18, flexWrap: "wrap" }}>
          <Stat label="Strategy" value={route.strategy}/>
          <Stat label="RPM" value={route.rpm.toLocaleString()}/>
          <Stat label="p95" value={`${route.p95}ms`}/>
          <Stat label="Error rate" value={`${route.errRate}%`}/>
          <Stat label="Attached keys" value={route.attached}/>
        </div>

        <h4 style={{ margin: "0 0 10px", fontSize: 13, fontWeight: 600, color: "var(--fg-muted)", textTransform: "uppercase", letterSpacing: "0.06em" }}>Routing diagram</h4>
        <RoutingViz route={route}/>

        <div style={{ height: 24 }}></div>
        <h4 style={{ margin: "0 0 10px", fontSize: 13, fontWeight: 600, color: "var(--fg-muted)", textTransform: "uppercase", letterSpacing: "0.06em" }}>Configuration</h4>
        <div className="code">
{`route: `}<span className="s">"{route.name}"</span>{`
strategy: `}<span className="k">{route.strategy}</span>{`
members:`}
{route.members.map(m => (
  <React.Fragment key={m.model}>{`
  - model: `}<span className="s">"{m.model}"</span>{`
    weight: `}<span className="n">{m.weight}</span>{`
    role: `}<span className="k">{m.role}</span></React.Fragment>
))}{`
budget:
  monthly_cap: `}<span className="n">5000</span>{`
  alert_at_pct: `}<span className="n">85</span>{`
rate_limit:
  rpm: `}<span className="n">600</span>{`
  tpm: `}<span className="n">200000</span>
        </div>
      </div>
    </div>
  );
}

function Stat({ label, value }) {
  return (
    <div style={{ minWidth: 100 }}>
      <div className="subtle" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>{label}</div>
      <div className="mono tnum" style={{ fontSize: 18, fontWeight: 600, marginTop: 2 }}>{value}</div>
    </div>
  );
}

// Routing visualizer — 1 source → N members
function RoutingViz({ route }) {
  const isFallback = route.strategy === "fallback";
  return (
    <div style={{ background: "var(--bg-sunken)", border: "1px solid var(--border)", borderRadius: 12, padding: 24, position: "relative", minHeight: 220 }}>
      <svg width="100%" height={Math.max(160, route.members.length * 56 + 40)} style={{ position: "absolute", inset: 0, padding: 24 }}>
        {route.members.map((m, i) => {
          const total = route.members.length;
          const y = 30 + (i * 56);
          const sourceY = 30 + (Math.max(0, total - 1) * 56) / 2;
          return (
            <path key={i}
              d={`M 140 ${sourceY} C 220 ${sourceY}, 240 ${y}, 320 ${y}`}
              stroke={isFallback && i > 0 ? "var(--warn)" : "var(--brand)"}
              strokeOpacity={isFallback && i > 0 ? 0.5 : 0.85}
              strokeDasharray={isFallback && i > 0 ? "4 4" : "none"}
              strokeWidth={2 + (m.weight / 100) * 1.2}
              fill="none"/>
          );
        })}
      </svg>
      <div style={{ position: "relative", display: "grid", gridTemplateColumns: "180px 1fr", gap: 100, alignItems: "center" }}>
        <div style={{ background: "var(--bg-elev)", border: "1px solid var(--border-strong)", borderRadius: 10, padding: 14 }}>
          <div className="subtle" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: "0.06em" }}>Route</div>
          <div className="mono" style={{ fontSize: 14, fontWeight: 600, marginTop: 4 }}>{route.name}</div>
          <div className="subtle mono" style={{ fontSize: 11, marginTop: 4 }}>{route.rpm.toLocaleString()} rpm</div>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {route.members.map((m, i) => (
            <div key={i} className="row between" style={{
              background: "var(--bg-elev)",
              border: "1px solid var(--border)",
              borderRadius: 10, padding: "10px 14px",
              borderLeft: `3px solid ${m.role === "fallback" ? "var(--warn)" : "var(--brand)"}`
            }}>
              <div className="row" style={{ gap: 10 }}>
                <Logo kind={m.logo}/>
                <div>
                  <div className="mono" style={{ fontWeight: 500, fontSize: 13 }}>{m.model}</div>
                  <div className="subtle" style={{ fontSize: 11 }}>{m.role}</div>
                </div>
              </div>
              <div className="row" style={{ gap: 12 }}>
                <span className="mono tnum subtle" style={{ fontSize: 12 }}>w {m.weight}</span>
                <Pill kind={m.role === "fallback" ? "warn" : "success"} icon>
                  {m.role === "fallback" ? "standby" : `${m.weight}%`}
                </Pill>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

window.LG.Providers = Providers;
window.LG.AddProviderDrawer = AddProviderDrawer;
window.LG.Routes = Routes;
