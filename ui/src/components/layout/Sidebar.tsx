"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { sidebarNav } from "./sidebar-config";

function NavIcon({ d }: { d: string }) {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="ico"
      aria-hidden
    >
      {d.split(" M").map((seg, i) => (
        <path key={i} d={i === 0 ? seg : `M${seg}`} />
      ))}
    </svg>
  );
}

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="sidebar">
      <div className="sb-brand">
        <div className="mark">G</div>
        <div className="name">
          LLMGopher <span className="ver">v0.4.2</span>
        </div>
      </div>

      <div className="sb-env" role="button" tabIndex={0} aria-label="Switch environment">
        <div>
          <div className="lbl">Environment</div>
          <div className="val">production</div>
        </div>
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden
        >
          <path d="M6 9l6 6 6-6" />
        </svg>
      </div>

      <nav style={{ flex: 1, overflowY: "auto", paddingBottom: 12 }} aria-label="Main">
        {sidebarNav.map((group) => (
          <div key={group.group}>
            <div className="sb-section">{group.group}</div>
            <div className="sb-nav">
              {group.items.map((item) => {
                const isActive =
                  item.href === pathname ||
                  (item.href !== "/" && pathname.startsWith(item.href));
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`sb-link${isActive ? " active" : ""}`}
                    aria-current={isActive ? "page" : undefined}
                    data-testid={`nav-${item.testId}`}
                  >
                    <NavIcon d={item.iconPath} />
                    <span>{item.name}</span>
                    {item.badge === "live" && (
                      <span
                        className="badge"
                        style={{ display: "inline-flex", alignItems: "center", gap: 4 }}
                      >
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
                    )}
                    {item.badge && item.badge !== "live" && (
                      <span className="badge">{item.badge}</span>
                    )}
                  </Link>
                );
              })}
            </div>
          </div>
        ))}
      </nav>

      <div className="sb-foot">
        <div className="av" aria-hidden>ED</div>
        <div className="who">
          <b>Elliott Davis</b>
          <span>Platform admin</span>
        </div>
        <button
          className="icon-btn"
          style={{ marginLeft: "auto" }}
          aria-label="Account settings"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.75"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden
          >
            <path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33h0a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51h0a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82v0a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
        </button>
      </div>
    </aside>
  );
}
