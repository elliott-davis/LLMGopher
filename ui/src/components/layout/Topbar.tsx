"use client";

import { usePathname } from "next/navigation";
import { getBreadcrumb } from "./sidebar-config";

export function Topbar() {
  const pathname = usePathname();
  const [group, page] = getBreadcrumb(pathname);

  return (
    <header className="topbar">
      <div className="crumbs" data-testid="page-title">
        <span>{group}</span>
        <span className="sep">/</span>
        <b>{page}</b>
      </div>

      <div className="search" role="search" aria-label="Jump to anything">
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
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
        <span>jump to anything…</span>
        <kbd>⌘K</kbd>
      </div>

      <button className="icon-btn" aria-label="Notifications">
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
          <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
          <path d="M13.73 21a2 2 0 0 1-3.46 0" />
        </svg>
      </button>

      <button className="icon-btn" aria-label="Help">
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
          <circle cx="12" cy="12" r="10" />
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
          <path d="M12 17h.01" />
        </svg>
      </button>
    </header>
  );
}
