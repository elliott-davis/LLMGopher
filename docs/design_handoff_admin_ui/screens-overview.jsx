/* global React, LG */
const { useState, useMemo, useEffect, useRef } = React;
const { Icon, Icons, Spark, Pill, StatusPill, Logo, Toggle } = LG;
const D = window.LG_DATA;

// ─────────── Overview ───────────
function Overview({ layout }) {
  return (
    <div className={`dash-${layout}`}>
      <div className="page-head">
        <div>
          <h1>Overview</h1>
          <div className="sub">Real-time traffic across all routes · last 24h · auto-refresh</div>
        </div>
        <div className="actions">
          <button className="btn"><Icon d={Icons.refresh}/>Refresh</button>
          <select className="btn" style={{ paddingRight: 8 }}>
            <option>Last 24h</option><option>Last 7d</option><option>Last 30d</option>
          </select>
          <button className="btn primary"><Icon d={Icons.download}/>Export</button>
        </div>
      </div>

      {layout !== "tables" && <KpiRow />}

      {layout === "charts" && <ChartsRow />}

      <div style={{ height: 16 }}></div>

      <FlowStrip />

      <div style={{ height: 16 }}></div>

      <div className="split split-2">
        <TopRoutesCard />
        <ProvidersHealthCard />
      </div>

      <div style={{ height: 16 }}></div>

      <div className="split split-2">
        <RecentActivityCard />
        <BudgetGlanceCard />
      </div>
    </div>
  );
}

function KpiRow() {
  const d = D.dash;
  const items = [
    { label: "Requests", value: fmt(d.requests24h), delta: `+${d.requestsDelta}%`, dir: "up", spark: d.requestsSpark, hint: "vs prev 24h" },
    { label: "Spend",    value: `$${d.spend24h.toFixed(2)}`, delta: `${d.spendDelta}%`, dir: "down", spark: d.spendSpark, hint: "vs prev 24h" },
    { label: "Error rate", value: `${d.errors24h}%`, delta: `${d.errorsDelta}%`, dir: "down", spark: d.errorsSpark, hint: "lower is better" },
    { label: "p95 latency", value: `${d.p95}ms`, delta: `+${d.p95Delta}%`, dir: "up", spark: d.p95Spark, hint: "across all routes" },
  ];
  return (
    <div className="kpi-grid">
      {items.map((k, i) => (
        <div key={i} className="kpi">
          <div className="label">{k.label}</div>
          <div className="value tnum">{k.value}</div>
          <div className="delta">
            <b className={k.dir === "down" && (k.label === "Spend" || k.label === "Error rate") ? "" : k.dir === "down" ? "down" : ""}>{k.delta}</b>
            <span className="subtle">{k.hint}</span>
          </div>
          <Spark data={k.spark} w={160} h={40} />
        </div>
      ))}
    </div>
  );
}

function ChartsRow() {
  return (
    <>
      <div style={{ height: 14 }}></div>
      <div className="split charts-row" style={{ gridTemplateColumns: "2fr 1fr" }}>
        <div className="card">
          <div className="card-head">
            <h3>Traffic by route</h3>
            <span className="sub">stacked · last 24h</span>
            <div className="right"><Pill>requests</Pill></div>
          </div>
          <div className="card-body">
            <StackedTrafficChart />
          </div>
        </div>
        <div className="card">
          <div className="card-head">
            <h3>Spend by team</h3>
            <span className="sub">today</span>
          </div>
          <div className="card-body">
            <SpendByTeam />
          </div>
        </div>
      </div>
    </>
  );
}

// Live request flow strip — middleware chain visualization
function FlowStrip() {
  const [pulses, setPulses] = useState([]);
  const stages = ["Ingress", "Auth", "Rate-limit", "Guardrails", "Route", "Provider"];
  useEffect(() => {
    let id = 1;
    const t = setInterval(() => {
      setPulses(p => [
        ...p.slice(-12),
        { id: id++, lane: Math.floor(Math.random() * 3), kind: Math.random() < 0.05 ? "err" : (Math.random() < 0.15 ? "fb" : "ok"), start: performance.now() }
      ]);
    }, 380);
    return () => clearInterval(t);
  }, []);
  return (
    <div className="card" style={{ overflow: "hidden" }}>
      <div className="card-head">
        <h3>Live request flow</h3>
        <span className="sub mono">middleware → upstream</span>
        <div className="right">
          <span className="mono subtle" style={{ fontSize: 11 }}>● 4,212 rpm</span>
          <Pill kind="success" icon>healthy</Pill>
        </div>
      </div>
      <div style={{ position: "relative", padding: "26px 22px 22px", background: "var(--bg-sunken)" }}>
        {/* stage labels */}
        <div style={{ display: "grid", gridTemplateColumns: `repeat(${stages.length}, 1fr)`, alignItems: "center", position: "relative" }}>
          {stages.map((s, i) => (
            <div key={s} style={{ textAlign: "center", position: "relative" }}>
              <div style={{
                width: 38, height: 38, borderRadius: 10,
                background: "var(--bg-elev)", border: "1px solid var(--border)",
                margin: "0 auto", display: "grid", placeItems: "center",
                color: "var(--fg-muted)", fontFamily: "var(--font-mono)", fontSize: 11,
                position: "relative", zIndex: 2
              }}>{i + 1}</div>
              <div style={{ marginTop: 6, fontSize: 11.5, color: "var(--fg-muted)", fontWeight: 500 }}>{s}</div>
            </div>
          ))}
          {/* connector line */}
          <div style={{
            position: "absolute", left: "8.33%", right: "8.33%",
            top: 19, height: 1, background: "var(--border)", zIndex: 1
          }}></div>
          {/* pulses */}
          {[0, 1, 2].map(lane => (
            <div key={lane} style={{
              position: "absolute", left: "8.33%", right: "8.33%",
              top: 19 + (lane - 1) * 6, height: 0, zIndex: 3
            }}>
              {pulses.filter(p => p.lane === lane).map(p => (
                <Pulse key={p.id} kind={p.kind} />
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function Pulse({ kind }) {
  const [x, setX] = useState(0);
  useEffect(() => {
    const start = performance.now();
    let raf;
    const tick = (t) => {
      const dt = t - start;
      const pct = Math.min(1, dt / 2200);
      setX(pct);
      if (pct < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, []);
  const color = kind === "err" ? "var(--danger)" : kind === "fb" ? "var(--warn)" : "var(--brand)";
  return (
    <div style={{
      position: "absolute",
      left: `${(x * 100).toFixed(2)}%`,
      transform: "translate(-50%, -50%)",
      width: 8, height: 8, borderRadius: "50%",
      background: color,
      boxShadow: `0 0 0 4px color-mix(in oklch, ${color} 22%, transparent)`,
      opacity: x < 1 ? 1 : 0,
      transition: "opacity 0.2s",
    }}></div>
  );
}

// Stacked traffic chart — 24 buckets, 4 routes
function StackedTrafficChart() {
  const buckets = 24;
  const colors = ["var(--brand)", "oklch(0.65 0.13 150)", "oklch(0.7 0.13 50)", "oklch(0.65 0.13 290)"];
  const series = useMemo(() => (
    [0, 1, 2, 3].map(s => Array.from({ length: buckets }, () => 200 + Math.random() * 800 + s * 200))
  ), []);
  const totals = Array.from({ length: buckets }, (_, i) => series.reduce((a, s) => a + s[i], 0));
  const max = Math.max(...totals);
  const w = 100;
  return (
    <div>
      <svg viewBox={`0 0 ${buckets * 16} 180`} width="100%" height="200" preserveAspectRatio="none">
        {Array.from({ length: buckets }, (_, i) => {
          let acc = 0;
          return series.map((s, si) => {
            const v = s[i];
            const h = (v / max) * 160;
            const y = 170 - acc - h;
            acc += h;
            return <rect key={si} x={i * 16 + 2} y={y} width="12" height={Math.max(1, h - 1)} fill={colors[si]} rx="2" />;
          });
        })}
      </svg>
      <div className="row" style={{ gap: 16, marginTop: 8, fontSize: 11.5 }}>
        {["chat-prod", "chat-cheap", "chat-long", "chat-eu"].map((n, i) => (
          <span key={n} className="row" style={{ gap: 5 }}>
            <span style={{ width: 9, height: 9, borderRadius: 2, background: colors[i] }}></span>
            <span className="mono">{n}</span>
          </span>
        ))}
      </div>
    </div>
  );
}

function SpendByTeam() {
  const teams = [...D.teams].sort((a,b) => b.budget - a.budget).slice(0,5);
  const max = Math.max(...teams.map(t => t.cap));
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
      {teams.map(t => {
        const pct = t.budget / t.cap;
        const warn = pct > 0.85;
        return (
          <div key={t.id}>
            <div className="row between" style={{ marginBottom: 4, fontSize: 12.5 }}>
              <span style={{ fontWeight: 500 }}>{t.name}</span>
              <span className="mono tnum subtle">${t.budget.toLocaleString()} <span style={{ opacity: 0.5 }}>/ ${t.cap.toLocaleString()}</span></span>
            </div>
            <div style={{ height: 6, background: "var(--bg-sunken)", borderRadius: 3, overflow: "hidden" }}>
              <div style={{ width: `${(t.cap/max)*100}%`, height: "100%", background: "var(--bg-active)", position: "relative" }}>
                <div style={{ width: `${pct*100}%`, height: "100%", background: warn ? "var(--warn)" : "var(--brand)" }}></div>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function TopRoutesCard() {
  return (
    <div className="card">
      <div className="card-head">
        <h3>Top routes</h3>
        <span className="sub">by request volume</span>
        <div className="right"><a className="btn ghost sm">View all <Icon d={Icons.arrowRight}/></a></div>
      </div>
      <div className="card-body tight">
        <table className="tbl">
          <thead><tr><th>Route</th><th>Strategy</th><th className="num">RPM</th><th className="num">p95</th><th className="num">Errors</th></tr></thead>
          <tbody>
            {[...D.routes].sort((a,b)=>b.rpm-a.rpm).map(r => (
              <tr key={r.id}>
                <td>
                  <div className="row" style={{ gap: 8 }}>
                    <div className="row" style={{ gap: -4 }}>
                      {r.members.slice(0,3).map((m,i)=>(
                        <span key={i} style={{ marginLeft: i?-6:0 }}><Logo kind={m.logo}/></span>
                      ))}
                    </div>
                    <span className="mono" style={{ fontWeight: 500 }}>{r.name}</span>
                  </div>
                </td>
                <td><Pill>{r.strategy}</Pill></td>
                <td className="num tnum">{r.rpm.toLocaleString()}</td>
                <td className="num tnum">{r.p95}ms</td>
                <td className="num tnum" style={{ color: r.errRate > 1 ? "var(--warn)" : "inherit" }}>{r.errRate}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ProvidersHealthCard() {
  return (
    <div className="card">
      <div className="card-head">
        <h3>Provider health</h3>
        <span className="sub">live</span>
        <div className="right"><a className="btn ghost sm">Manage <Icon d={Icons.arrowRight}/></a></div>
      </div>
      <div className="card-body" style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {D.providers.map(p => (
          <div key={p.id} className="row between" style={{ padding: "8px 4px", borderBottom: "1px solid var(--border)" }}>
            <div className="row" style={{ gap: 10 }}>
              <Logo kind={p.logo}/>
              <div>
                <div style={{ fontWeight: 500, fontSize: 13 }}>{p.name}</div>
                <div className="subtle mono" style={{ fontSize: 11 }}>{p.region} · {p.deployments} deployments</div>
              </div>
            </div>
            <div className="row" style={{ gap: 14 }}>
              <span className="mono tnum subtle" style={{ fontSize: 11.5 }}>p95 {p.latencyP95}ms</span>
              <StatusPill status={p.status}/>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function RecentActivityCard() {
  return (
    <div className="card">
      <div className="card-head">
        <h3>Recent requests</h3>
        <span className="sub">streaming</span>
        <div className="right"><Pill icon kind="success">live</Pill></div>
      </div>
      <div className="card-body tight" style={{ maxHeight: 320, overflow: "auto" }}>
        <table className="tbl">
          <tbody>
            {D.requests.slice(0, 8).map(r => (
              <tr key={r.id}>
                <td className="id" style={{ width: 90 }}>{r.t}</td>
                <td><Logo kind={getProviderKind(r.model)}/></td>
                <td><span className="mono" style={{ fontSize: 12 }}>{r.model}</span></td>
                <td><Pill>{r.route}</Pill></td>
                <td>
                  {r.status === 200
                    ? <Pill kind="success">{r.status}</Pill>
                    : r.status >= 500
                      ? <Pill kind="danger">{r.status}</Pill>
                      : <Pill kind="warn">{r.status}</Pill>}
                </td>
                <td className="num tnum subtle">{r.latency}ms</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function BudgetGlanceCard() {
  const total = D.teams.reduce((a,t)=>a+t.budget,0);
  const cap   = D.teams.reduce((a,t)=>a+t.cap,0);
  return (
    <div className="card">
      <div className="card-head">
        <h3>Org budget</h3>
        <span className="sub">monthly</span>
        <div className="right"><a className="btn ghost sm">Budgets <Icon d={Icons.arrowRight}/></a></div>
      </div>
      <div className="card-body">
        <div className="row between">
          <div>
            <div className="value mono tnum" style={{ fontSize: 28, fontWeight: 600, letterSpacing: "-0.02em" }}>${total.toLocaleString()}</div>
            <div className="subtle" style={{ fontSize: 12 }}>of ${cap.toLocaleString()} cap · {((total/cap)*100).toFixed(1)}% used</div>
          </div>
          <BudgetRing pct={total/cap} />
        </div>
        <div style={{ height: 14 }}></div>
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {D.teams.slice(0,4).map(t => {
            const pct = t.budget / t.cap;
            return (
              <div key={t.id}>
                <div className="row between" style={{ fontSize: 12.5, marginBottom: 4 }}>
                  <span>{t.name}</span>
                  <span className="mono subtle tnum">${t.budget.toLocaleString()} / ${t.cap.toLocaleString()}</span>
                </div>
                <div style={{ height: 4, background: "var(--bg-sunken)", borderRadius: 2 }}>
                  <div style={{ width: `${pct*100}%`, height: "100%", background: pct > 0.85 ? "var(--warn)" : "var(--brand)", borderRadius: 2 }}></div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function BudgetRing({ pct, size = 64 }) {
  const r = (size - 8) / 2;
  const c = 2 * Math.PI * r;
  return (
    <svg width={size} height={size}>
      <circle cx={size/2} cy={size/2} r={r} stroke="var(--bg-sunken)" strokeWidth="6" fill="none"/>
      <circle cx={size/2} cy={size/2} r={r}
        stroke={pct > 0.85 ? "var(--warn)" : "var(--brand)"}
        strokeWidth="6" fill="none" strokeLinecap="round"
        strokeDasharray={`${c * pct} ${c}`}
        transform={`rotate(-90 ${size/2} ${size/2})`}/>
      <text x="50%" y="52%" textAnchor="middle" fontSize="11" fill="var(--fg)" fontFamily="var(--font-mono)" fontWeight="600">
        {Math.round(pct*100)}%
      </text>
    </svg>
  );
}

function fmt(n) {
  if (n >= 1e6) return (n/1e6).toFixed(2) + "M";
  if (n >= 1e3) return (n/1e3).toFixed(1) + "K";
  return n.toString();
}

function getProviderKind(model) {
  if (model.includes("gpt")) return "openai";
  if (model.includes("claude")) return "anthropic";
  if (model.includes("gemini")) return "google";
  if (model.includes("mistral")) return "mistral";
  if (model.includes("llama") && !model.includes("vllm")) return "bedrock";
  if (model.includes("vllm")) return "vllm";
  return "openai";
}

window.LG.Overview = Overview;
window.LG.getProviderKind = getProviderKind;
window.LG.fmt = fmt;
