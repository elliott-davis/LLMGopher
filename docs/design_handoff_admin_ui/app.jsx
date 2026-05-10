/* global React, ReactDOM, LG */
const { useState, useEffect } = React;
const { Icon, Icons } = LG;

const NAV = [
  { group: "Operate", items: [
    { id: "overview",   label: "Overview",      icon: Icons.overview },
    { id: "logs",       label: "Logs",          icon: Icons.logs, badge: "live" },
    { id: "audit",      label: "Audit log",     icon: Icons.audit },
  ]},
  { group: "Configure", items: [
    { id: "providers",  label: "Providers",     icon: Icons.providers, badge: "6" },
    { id: "routes",     label: "Routes",        icon: Icons.routes,    badge: "5" },
    { id: "guardrails", label: "Guardrails",    icon: Icons.guardrails },
  ]},
  { group: "Govern", items: [
    { id: "keys",       label: "API Keys",      icon: Icons.keys, badge: "7" },
    { id: "teams",      label: "Teams",         icon: Icons.teams, badge: "6" },
    { id: "budgets",    label: "Budgets",       icon: Icons.budgets },
    { id: "rate",       label: "Rate limits",   icon: Icons.rate },
  ]},
  { group: "Org", items: [
    { id: "settings",   label: "Settings",      icon: Icons.settings },
  ]},
];

function App() {
  const [route, setRoute] = useState("overview");
  const [addProviderOpen, setAddProviderOpen] = useState(false);
  const [keyDrawer, setKeyDrawer] = useState(null);
  const [reqDrawer, setReqDrawer] = useState(null);

  // Tweaks (defaults persisted via host edit-mode)
  const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
    "theme": "light",
    "density": "comfortable",
    "dashboardLayout": "cards"
  }/*EDITMODE-END*/;
  const [tweaks, setTweak] = window.useTweaks(TWEAK_DEFAULTS);

  // Apply theme + density to <html>
  useEffect(() => {
    document.documentElement.dataset.theme = tweaks.theme;
    document.documentElement.dataset.density = tweaks.density;
  }, [tweaks.theme, tweaks.density]);

  const crumbs = (() => {
    const map = {
      overview: ["Operate", "Overview"],
      logs: ["Operate", "Logs"],
      audit: ["Operate", "Audit log"],
      providers: ["Configure", "Providers"],
      routes: ["Configure", "Routes"],
      guardrails: ["Configure", "Guardrails"],
      keys: ["Govern", "API Keys"],
      teams: ["Govern", "Teams"],
      budgets: ["Govern", "Budgets"],
      rate: ["Govern", "Rate limits"],
      settings: ["Org", "Settings"],
    };
    return map[route] || ["", ""];
  })();

  return (
    <div className="app">
      <Sidebar route={route} setRoute={setRoute}/>
      <div className="main">
        <Topbar crumbs={crumbs}/>
        <div className="content" data-screen-label={crumbs[1]}>
          {route === "overview"   && <LG.Overview layout={tweaks.dashboardLayout}/>}
          {route === "providers"  && <LG.Providers onAdd={()=>setAddProviderOpen(true)}/>}
          {route === "routes"     && <LG.Routes/>}
          {route === "keys"       && <LG.Keys onInspect={setKeyDrawer}/>}
          {route === "teams"      && <LG.Teams/>}
          {route === "budgets"    && <LG.Budgets/>}
          {route === "rate"       && <LG.RateLimits/>}
          {route === "logs"       && <LG.Logs onInspect={setReqDrawer}/>}
          {route === "guardrails" && <LG.Guardrails/>}
          {route === "audit"      && <LG.Audit/>}
          {route === "settings"   && <LG.Settings/>}
        </div>
      </div>
      {addProviderOpen && <LG.AddProviderDrawer open={addProviderOpen} onClose={()=>setAddProviderOpen(false)}/>}
      {keyDrawer && <LG.KeyDrawer keyData={keyDrawer} onClose={()=>setKeyDrawer(null)}/>}
      {reqDrawer && <LG.RequestInspector req={reqDrawer} onClose={()=>setReqDrawer(null)}/>}
      <TweaksUI tweaks={tweaks} setTweak={setTweak}/>
    </div>
  );
}

function Sidebar({ route, setRoute }) {
  return (
    <aside className="sidebar">
      <div className="sb-brand">
        <div className="mark">G</div>
        <div className="name">LLMGopher <span className="ver">v0.4.2</span></div>
      </div>
      <div className="sb-env">
        <div>
          <div className="lbl">Environment</div>
          <div className="val">production</div>
        </div>
        <Icon d="M6 9l6 6 6-6" size={14}/>
      </div>
      <nav style={{ flex: 1, overflowY: "auto", paddingBottom: 12 }}>
        {NAV.map(group => (
          <div key={group.group}>
            <div className="sb-section">{group.group}</div>
            <div className="sb-nav">
              {group.items.map(it => (
                <a key={it.id} className={`sb-link ${route === it.id ? "active" : ""}`} onClick={()=>setRoute(it.id)}>
                  <Icon d={it.icon}/>
                  <span>{it.label}</span>
                  {it.badge && (
                    it.badge === "live"
                      ? <span className="badge" style={{ display: "inline-flex", alignItems: "center", gap: 4 }}>
                          <span style={{ width: 6, height: 6, borderRadius: "50%", background: "var(--success)", boxShadow: "0 0 0 3px color-mix(in oklch, var(--success) 25%, transparent)" }}></span>
                          live
                        </span>
                      : <span className="badge">{it.badge}</span>
                  )}
                </a>
              ))}
            </div>
          </div>
        ))}
      </nav>
      <div className="sb-foot">
        <div className="av">ED</div>
        <div className="who"><b>Elliott Davis</b><span>Platform admin</span></div>
        <button className="icon-btn" style={{ marginLeft: "auto" }}><Icon d={Icons.settings} size={14}/></button>
      </div>
    </aside>
  );
}

function Topbar({ crumbs }) {
  return (
    <header className="topbar">
      <div className="crumbs">
        <span>{crumbs[0]}</span>
        <span className="sep">/</span>
        <b>{crumbs[1]}</b>
      </div>
      <div className="search">
        <Icon d={Icons.search} size={14}/>
        <span>jump to anything…</span>
        <kbd>⌘K</kbd>
      </div>
      <button className="icon-btn"><Icon d={Icons.bell} size={14}/></button>
      <button className="icon-btn"><Icon d="M12 17v.01 M12 13.5a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3z M5 9.5a7 7 0 1 1 14 0c0 3.5-2 5-3 6.5a3 3 0 0 0-.5 1.5v.5a2 2 0 0 1-2 2h-3a2 2 0 0 1-2-2v-.5a3 3 0 0 0-.5-1.5C7 14.5 5 13 5 9.5z" size={14}/></button>
    </header>
  );
}

function TweaksUI({ tweaks, setTweak }) {
  const { TweaksPanel, TweakSection, TweakRadio, TweakSelect } = window;
  return (
    <TweaksPanel title="Tweaks">
      <TweakSection label="Appearance">
        <TweakRadio
          label="Theme"
          value={tweaks.theme}
          options={[{ value: "light", label: "Light" }, { value: "dark", label: "Dark" }]}
          onChange={v => setTweak("theme", v)}/>
        <TweakRadio
          label="Density"
          value={tweaks.density}
          options={[{ value: "comfortable", label: "Comfy" }, { value: "compact", label: "Compact" }]}
          onChange={v => setTweak("density", v)}/>
      </TweakSection>
      <TweakSection label="Overview layout">
        <TweakSelect
          label="Dashboard"
          value={tweaks.dashboardLayout}
          options={[
            { value: "cards", label: "Cards-first (KPIs + summaries)" },
            { value: "charts", label: "Charts-first (stacked traffic)" },
            { value: "tables", label: "Tables-first (no KPIs)" },
          ]}
          onChange={v => setTweak("dashboardLayout", v)}/>
      </TweakSection>
    </TweaksPanel>
  );
}

const root = ReactDOM.createRoot(document.getElementById("root"));
root.render(<App/>);
