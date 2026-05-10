/* global React */
const { useState, useMemo, useEffect, useRef } = React;

// Tiny icon set (lucide-style strokes)
const Icon = ({ d, size = 16 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" className="ico"><path d={d} /></svg>
);
const Icons = {
  overview: "M3 13h8V3H3zM13 21h8V11h-8zM3 21h8v-6H3zM13 9h8V3h-8z",
  providers: "M4 7h16M4 12h16M4 17h16 M8 7v10 M16 7v10",
  models: "M12 2 4 7v10l8 5 8-5V7zM12 2v20 M4 7l8 5 8-5",
  routes: "M3 6h6a4 4 0 0 1 4 4v4a4 4 0 0 0 4 4h4 M17 14l4 4-4 4",
  keys: "M21 2 13 10 m-2 2a3 3 0 1 1-4 4 3 3 0 0 1 4-4z m4-6 3 3 m-5-1 3 3",
  teams: "M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2 M22 21v-2a4 4 0 0 0-3-3.87 M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z M16 3.13a4 4 0 0 1 0 7.75",
  budgets: "M12 2v20 M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6",
  rate: "M12 8v4l3 2 M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z",
  logs: "M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M9 13h6 M9 17h6",
  guardrails: "M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z",
  audit: "M9 11l3 3L22 4 M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11",
  settings: "M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33h0a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51h0a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82v0a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z",
  search: "M21 21l-4.3-4.3 M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16z",
  bell: "M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9 M13.73 21a2 2 0 0 1-3.46 0",
  plus: "M12 5v14 M5 12h14",
  filter: "M22 3H2l8 9.46V19l4 2v-8.54z",
  more: "M12 12a1 1 0 1 0 0-2 1 1 0 0 0 0 2z M19 12a1 1 0 1 0 0-2 1 1 0 0 0 0 2z M5 12a1 1 0 1 0 0-2 1 1 0 0 0 0 2z",
  arrowRight: "M5 12h14 M12 5l7 7-7 7",
  copy: "M20 9h-9a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h9a2 2 0 0 0 2-2v-9a2 2 0 0 0-2-2z M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1",
  refresh: "M3 12a9 9 0 0 1 15-6.7L21 8 M21 3v5h-5 M21 12a9 9 0 0 1-15 6.7L3 16 M3 21v-5h5",
  close: "M18 6 6 18 M6 6l12 12",
  zap: "M13 2 3 14h9l-1 8 10-12h-9z",
  download: "M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4 M7 10l5 5 5-5 M12 15V3",
  shield: "M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z",
};

// Sparkline helper
function Spark({ data, w = 120, h = 28, color = "var(--brand)", filled = true }) {
  if (!data || !data.length) return null;
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = Math.max(max - min, 1e-3);
  const pts = data.map((v, i) => [(i / (data.length - 1)) * w, h - ((v - min) / range) * (h - 2) - 1]);
  const d = "M" + pts.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(" L ");
  const dArea = `M${pts[0][0]},${h} L ` + pts.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(" L ") + ` L ${pts[pts.length - 1][0]},${h} Z`;
  return (
    <svg width={w} height={h} className="spark">
      {filled && <path d={dArea} fill={`color-mix(in oklch, ${color} 18%, transparent)`} />}
      <path d={d} fill="none" stroke={color} strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" />
    </svg>
  );
}

// Pill
const Pill = ({ kind = "", icon, children }) => (
  <span className={`pill ${kind}`}>
    {icon && <span className="dot"></span>}
    {children}
  </span>
);

const StatusPill = ({ status }) => {
  const map = {
    healthy: { kind: "success", label: "healthy" },
    degraded: { kind: "warn", label: "degraded" },
    offline: { kind: "danger", label: "offline" },
    active: { kind: "success", label: "active" },
    throttled: { kind: "warn", label: "throttled" },
    warning: { kind: "warn", label: "near cap" },
    disabled: { kind: "", label: "disabled" },
  };
  const m = map[status] || { kind: "", label: status };
  return <Pill kind={m.kind} icon>{m.label}</Pill>;
};

const Logo = ({ kind, size = "" }) => {
  const letter = { openai: "O", anthropic: "A", google: "G", mistral: "M", bedrock: "B", vllm: "v" }[kind] || "·";
  return <span className={`logo ${kind} ${size}`}>{letter}</span>;
};

// Toggle
const Toggle = ({ on, onChange }) => (
  <span className={`toggle ${on ? "on" : ""}`} onClick={() => onChange(!on)} role="switch" aria-checked={on}></span>
);

window.LG = { Icon, Icons, Spark, Pill, StatusPill, Logo, Toggle };
