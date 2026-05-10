/* global React, LG */
const { useState, useMemo } = React;
const { Icon, Icons, Pill, StatusPill, Logo, Toggle } = LG;
const D3 = window.LG_DATA;

// ─────────── API Keys ───────────
function Keys({ onInspect }) {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>API Keys</h1>
          <div className="sub">Virtual keys with their own budget, rate-limit, and route allowlist. Keys are stored as <code className="mono">KEY:ID</code> and never echoed in logs.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.download}/>Export</button>
          <button className="btn primary"><Icon d={Icons.plus}/>New key</button>
        </div>
      </div>

      <div className="card">
        <div className="card-head">
          <h3>Active keys</h3>
          <span className="sub">{D3.keys.length} total</span>
          <div className="right">
            <span className="search" style={{ width: 240 }}><Icon d={Icons.search}/>name, prefix, team…</span>
          </div>
        </div>
        <div className="card-body tight">
          <table className="tbl">
            <thead>
              <tr>
                <th>Name</th><th>Key prefix</th><th>Team</th>
                <th>Budget (mo)</th><th className="num">RPM cap</th>
                <th>Status</th><th>Last used</th><th></th>
              </tr>
            </thead>
            <tbody>
              {D3.keys.map(k => {
                const pct = k.budget / k.budgetCap;
                return (
                  <tr key={k.id} onClick={()=>onInspect(k)} style={{ cursor: "pointer" }}>
                    <td>
                      <div style={{ fontWeight: 500 }}>{k.name}</div>
                      <div className="id" style={{ fontSize: 11 }}>{k.id} · created {k.created}</div>
                    </td>
                    <td className="mono" style={{ fontSize: 12, color: "var(--fg-muted)" }}>{k.prefix}</td>
                    <td><Pill>{k.team}</Pill></td>
                    <td>
                      <div className="row" style={{ gap: 10 }}>
                        <div style={{ width: 90, height: 5, background: "var(--bg-sunken)", borderRadius: 3 }}>
                          <div style={{
                            width: `${Math.min(100, pct*100)}%`, height: "100%",
                            background: pct > 0.85 ? "var(--warn)" : "var(--brand)",
                            borderRadius: 3,
                          }}></div>
                        </div>
                        <span className="mono tnum subtle" style={{ fontSize: 11.5 }}>${k.budget}/{k.budgetCap}</span>
                      </div>
                    </td>
                    <td className="num tnum">{k.rpm}</td>
                    <td><StatusPill status={k.status}/></td>
                    <td className="subtle" style={{ fontSize: 12 }}>{k.lastUsed}</td>
                    <td><button className="icon-btn" style={{ width: 26, height: 26 }} onClick={(e)=>{e.stopPropagation();}}><Icon d={Icons.more}/></button></td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function KeyDrawer({ keyData, onClose }) {
  if (!keyData) return null;
  const [tab, setTab] = useState("overview");
  const k = keyData;
  return (
    <div className="drawer-mask" onClick={onClose}>
      <div className="drawer" onClick={e=>e.stopPropagation()}>
        <div className="card-head">
          <div>
            <h3>{k.name}</h3>
            <div className="sub mono">{k.prefix}  ·  {k.id}</div>
          </div>
          <div className="right">
            <button className="btn sm"><Icon d={Icons.refresh}/>Rotate</button>
            <button className="btn sm danger">Disable</button>
            <button className="icon-btn" onClick={onClose}><Icon d={Icons.close}/></button>
          </div>
        </div>
        <div className="tabs">
          {["overview","routes","limits","activity"].map(t=>(
            <button key={t} className={`tab ${tab===t?"active":""}`} onClick={()=>setTab(t)}>{t}</button>
          ))}
        </div>
        <div style={{ flex: 1, overflow: "auto", padding: 22 }}>
          {tab === "overview" && (
            <>
              <div className="kpi-grid" style={{ gridTemplateColumns: "repeat(3, 1fr)" }}>
                <Stat3 label="Budget used" value={`$${k.budget}`} sub={`of $${k.budgetCap}`}/>
                <Stat3 label="Rate limit" value={`${k.rpm} rpm`} sub="per minute"/>
                <Stat3 label="Status" value={<StatusPill status={k.status}/>} sub={`last used ${k.lastUsed}`}/>
              </div>
              <div style={{ height: 18 }}></div>
              <h4 style={hSty}>Allowed routes</h4>
              <div className="row" style={{ gap: 6, flexWrap: "wrap" }}>
                {["chat-prod","chat-cheap","embeddings"].map(r => <Pill key={r} kind="brand">{r}</Pill>)}
              </div>
              <div style={{ height: 18 }}></div>
              <h4 style={hSty}>Spend last 7d</h4>
              <MiniBars data={[12, 18, 14, 22, 19, 28, 24]} max={30}/>
            </>
          )}
          {tab === "routes" && (
            <>
              <h4 style={hSty}>Route allowlist</h4>
              <table className="tbl">
                <thead><tr><th></th><th>Route</th><th>Strategy</th><th className="num">RPM allowed</th></tr></thead>
                <tbody>
                  {D3.routes.map(r => (
                    <tr key={r.id}>
                      <td><input type="checkbox" defaultChecked={["chat-prod","chat-cheap","embeddings"].includes(r.id)}/></td>
                      <td className="mono">{r.name}</td>
                      <td><Pill>{r.strategy}</Pill></td>
                      <td className="num tnum">{r.rpm > 1000 ? "no override" : "200"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}
          {tab === "limits" && (
            <div style={{ display: "grid", gap: 14 }}>
              <div className="field"><label>Requests per minute</label><input className="mono tnum" defaultValue={k.rpm}/></div>
              <div className="field"><label>Tokens per minute</label><input className="mono tnum" defaultValue="200000"/></div>
              <div className="field"><label>Monthly budget cap (USD)</label><input className="mono tnum" defaultValue={k.budgetCap}/></div>
              <div className="field"><label>Alert at % of cap</label><input className="mono tnum" defaultValue="85"/></div>
              <div className="row between" style={{ padding: 12, border: "1px solid var(--border)", borderRadius: 10, background: "var(--bg-sunken)" }}>
                <div>
                  <div style={{ fontWeight: 500 }}>Hard cap</div>
                  <div className="subtle" style={{ fontSize: 12 }}>Reject requests once budget cap is hit</div>
                </div>
                <Toggle on={true} onChange={()=>{}}/>
              </div>
            </div>
          )}
          {tab === "activity" && (
            <table className="tbl">
              <thead><tr><th>Time</th><th>Route</th><th>Model</th><th className="num">Tokens</th><th className="num">Cost</th></tr></thead>
              <tbody>
                {D3.requests.slice(0,8).map(r => (
                  <tr key={r.id}>
                    <td className="id">{r.t}</td>
                    <td><Pill>{r.route}</Pill></td>
                    <td className="mono" style={{ fontSize: 12 }}>{r.model}</td>
                    <td className="num tnum">{(r.tokens.in + r.tokens.out).toLocaleString()}</td>
                    <td className="num tnum">${r.cost.toFixed(4)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
const hSty = { margin: "0 0 10px", fontSize: 12, fontWeight: 600, color: "var(--fg-muted)", textTransform: "uppercase", letterSpacing: "0.06em" };

function Stat3({ label, value, sub }) {
  return (
    <div className="kpi" style={{ padding: 14 }}>
      <div className="label">{label}</div>
      <div className="value tnum" style={{ fontSize: 20 }}>{value}</div>
      <div className="delta"><span className="subtle">{sub}</span></div>
    </div>
  );
}

function MiniBars({ data, max }) {
  return (
    <div style={{ display: "flex", gap: 6, alignItems: "flex-end", height: 80 }}>
      {data.map((v, i) => (
        <div key={i} style={{ flex: 1, height: `${(v/max)*100}%`, background: "var(--brand)", borderRadius: 3, opacity: 0.85 }} title={`$${v}`}></div>
      ))}
    </div>
  );
}

// ─────────── Teams ───────────
function Teams() {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Teams</h1>
          <div className="sub">Group members and assign budgets, route access, and default rate limits at the team level.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.download}/>Sync from SSO</button>
          <button className="btn primary"><Icon d={Icons.plus}/>New team</button>
        </div>
      </div>

      <div className="split split-3">
        {D3.teams.map(t => {
          const pct = t.budget / t.cap;
          return (
            <div key={t.id} className="card" style={{ padding: 0 }}>
              <div style={{ padding: "16px 18px 10px", borderBottom: "1px solid var(--border)" }}>
                <div className="row between">
                  <h3 style={{ margin: 0, fontSize: 15, fontWeight: 600 }}>{t.name}</h3>
                  <Pill kind={pct > 0.85 ? "warn" : "success"}>{pct > 0.85 ? "near cap" : "active"}</Pill>
                </div>
                <div className="subtle mono" style={{ fontSize: 11, marginTop: 4 }}>{t.id} · owner {t.owner}</div>
              </div>
              <div style={{ padding: "12px 18px" }}>
                <div className="row between" style={{ marginBottom: 10 }}>
                  <span className="subtle" style={{ fontSize: 11.5, textTransform: "uppercase", letterSpacing: "0.06em" }}>Budget</span>
                  <span className="mono tnum" style={{ fontSize: 12 }}>${t.budget.toLocaleString()} <span className="subtle">/ ${t.cap.toLocaleString()}</span></span>
                </div>
                <div style={{ height: 6, background: "var(--bg-sunken)", borderRadius: 3, marginBottom: 14 }}>
                  <div style={{ width: `${pct*100}%`, height: "100%", background: pct > 0.85 ? "var(--warn)" : "var(--brand)", borderRadius: 3 }}></div>
                </div>
                <div className="row between">
                  <div>
                    <div className="value tnum" style={{ fontSize: 18, fontWeight: 600 }}>{t.members}</div>
                    <div className="subtle" style={{ fontSize: 11 }}>members</div>
                  </div>
                  <div>
                    <div className="value tnum" style={{ fontSize: 18, fontWeight: 600 }}>{t.keys}</div>
                    <div className="subtle" style={{ fontSize: 11 }}>keys</div>
                  </div>
                  <div>
                    <div className="value tnum" style={{ fontSize: 18, fontWeight: 600 }}>{t.models.length}</div>
                    <div className="subtle" style={{ fontSize: 11 }}>routes</div>
                  </div>
                </div>
              </div>
              <div style={{ borderTop: "1px solid var(--border)", padding: "10px 18px", display: "flex", gap: 4, flexWrap: "wrap" }}>
                {t.models.map(m => <Pill key={m} kind="brand">{m}</Pill>)}
              </div>
            </div>
          );
        })}
      </div>
    </>
  );
}

// ─────────── Rate Limits ───────────
function RateLimits() {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Rate limits & quotas</h1>
          <div className="sub">Apply RPM and TPM ceilings at the team, key, model, or route level. Ranks: key &gt; team &gt; route &gt; org default.</div>
        </div>
        <div className="actions">
          <button className="btn primary"><Icon d={Icons.plus}/>New rule</button>
        </div>
      </div>

      <div className="kpi-grid" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
        <Stat3 label="Active rules" value="42" sub="14 keys · 8 teams · 6 routes"/>
        <Stat3 label="Throttle events 24h" value="318" sub={<span style={{ color: "var(--warn)" }}>+12% vs prev day</span>}/>
        <Stat3 label="Top throttled key" value="ml-research" sub="88 trips today"/>
        <Stat3 label="Backend" value="redis://prod-cache" sub="< 1ms p99"/>
      </div>
      <div style={{ height: 18 }}></div>

      <div className="card">
        <div className="card-head"><h3>Rules</h3><div className="right"><Pill kind="brand">42 rules</Pill></div></div>
        <div className="card-body tight">
          <table className="tbl">
            <thead>
              <tr><th>Scope</th><th>Target</th><th className="num">RPM</th><th className="num">TPM</th><th>Burst</th><th>Window</th><th>Action</th><th>Status</th></tr>
            </thead>
            <tbody>
              {[
                ["org","* (default)","2000","500k","2x","60s","throttle","active"],
                ["team","Research","1500","400k","1.5x","60s","throttle","active"],
                ["team","Growth","800","100k","1x","60s","reject","active"],
                ["key","ml-research","1200","300k","1x","60s","throttle","tripped"],
                ["route","chat-prod","8000","1.5M","2x","60s","throttle","active"],
                ["route","chat-long","50","100k","1x","60s","reject","active"],
                ["model","gpt-4o","10000","800k","2x","60s","queue","active"],
                ["model","claude-sonnet-4-5","4000","400k","1x","60s","queue","active"],
              ].map((r,i) => (
                <tr key={i}>
                  <td><Pill>{r[0]}</Pill></td>
                  <td className="mono" style={{ fontSize: 12.5, fontWeight: 500 }}>{r[1]}</td>
                  <td className="num tnum">{r[2]}</td>
                  <td className="num tnum">{r[3]}</td>
                  <td className="mono tnum subtle" style={{ fontSize: 12 }}>{r[4]}</td>
                  <td className="mono subtle" style={{ fontSize: 12 }}>{r[5]}</td>
                  <td className="mono" style={{ fontSize: 12 }}>{r[6]}</td>
                  <td>{r[7] === "tripped" ? <Pill kind="warn" icon>tripped</Pill> : <Pill kind="success" icon>active</Pill>}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

// ─────────── Budgets ───────────
function Budgets() {
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Budgets & spend</h1>
          <div className="sub">Track spend by team, key, and model. Set hard and soft caps with notification thresholds.</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.download}/>Export CSV</button>
          <button className="btn primary"><Icon d={Icons.plus}/>New budget</button>
        </div>
      </div>

      <div className="kpi-grid" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
        <Stat3 label="Spend MTD" value="$28,422" sub={<span style={{ color: "var(--success)" }}>−2.1% vs last month</span>}/>
        <Stat3 label="Org cap" value="$45,000" sub="63% used"/>
        <Stat3 label="Forecast EOM" value="$41,200" sub="on track"/>
        <Stat3 label="Top spender" value="Research" sub="$9,870 · 80% of cap"/>
      </div>
      <div style={{ height: 18 }}></div>

      <div className="split split-2">
        <div className="card">
          <div className="card-head"><h3>By team</h3><span className="sub">monthly</span></div>
          <div className="card-body">
            {D3.teams.map(t => {
              const pct = t.budget / t.cap;
              const warn = pct > 0.85;
              return (
                <div key={t.id} style={{ marginBottom: 14 }}>
                  <div className="row between" style={{ marginBottom: 4 }}>
                    <span style={{ fontWeight: 500 }}>{t.name}</span>
                    <span className="mono tnum" style={{ fontSize: 12 }}>
                      ${t.budget.toLocaleString()} <span className="subtle">/ ${t.cap.toLocaleString()}</span>
                      {warn && <span style={{ color: "var(--warn)", marginLeft: 8 }}>⚠</span>}
                    </span>
                  </div>
                  <div style={{ height: 8, background: "var(--bg-sunken)", borderRadius: 4 }}>
                    <div style={{ width: `${pct*100}%`, height: "100%", background: warn ? "var(--warn)" : "var(--brand)", borderRadius: 4 }}></div>
                    <div style={{ width: "85%", borderRight: "1px dashed var(--fg-subtle)", height: 12, marginTop: -10, opacity: 0.5 }}></div>
                  </div>
                  <div className="subtle" style={{ fontSize: 11, marginTop: 3 }}>alert at 85% · hard cap at 100%</div>
                </div>
              );
            })}
          </div>
        </div>

        <div className="card">
          <div className="card-head"><h3>By model</h3><span className="sub">last 30d</span></div>
          <div className="card-body tight">
            <table className="tbl">
              <thead><tr><th>Model</th><th className="num">Requests</th><th className="num">Tokens</th><th className="num">Cost</th></tr></thead>
              <tbody>
                {[
                  ["claude-sonnet-4.5","anthropic", 412082, "1.42B", 9420.12],
                  ["gpt-4o","openai", 312441, "880M", 6240.88],
                  ["gpt-4o-mini","openai", 1204022, "2.1B", 1080.40],
                  ["gemini-2.5-flash","google", 880222, "1.6B", 720.30],
                  ["claude-haiku-4.5","anthropic", 502912, "1.0B", 612.50],
                  ["mistral-large-2","mistral", 88400, "210M", 412.20],
                  ["llama-3.3-70b","bedrock", 122322, "440M", 318.10],
                ].map((r,i) => (
                  <tr key={i}>
                    <td>
                      <div className="row" style={{ gap: 8 }}>
                        <Logo kind={r[1]}/>
                        <span className="mono" style={{ fontSize: 12.5 }}>{r[0]}</span>
                      </div>
                    </td>
                    <td className="num tnum">{r[2].toLocaleString()}</td>
                    <td className="num tnum">{r[3]}</td>
                    <td className="num tnum">${r[4].toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </>
  );
}

window.LG.Keys = Keys;
window.LG.KeyDrawer = KeyDrawer;
window.LG.Teams = Teams;
window.LG.RateLimits = RateLimits;
window.LG.Budgets = Budgets;
