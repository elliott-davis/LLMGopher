"use client";

import { useEffect } from "react";

interface DrawerProps {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
}

export function Drawer({ open, onClose, children }: DrawerProps) {
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className="drawer-mask"
      role="dialog"
      aria-modal="true"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="drawer">{children}</div>
    </div>
  );
}

interface DrawerHeaderProps {
  title: string;
  onClose: () => void;
  actions?: React.ReactNode;
}

export function DrawerHeader({ title, onClose, actions }: DrawerHeaderProps) {
  return (
    <div
      className="card-head"
      style={{ padding: "14px 18px", flexShrink: 0 }}
    >
      <h3 style={{ margin: 0 }}>{title}</h3>
      {actions && <div className="right">{actions}</div>}
      <button
        className="icon-btn"
        onClick={onClose}
        aria-label="Close"
        style={{ marginLeft: actions ? 0 : "auto" }}
      >
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
          <path d="M18 6 6 18M6 6l12 12" />
        </svg>
      </button>
    </div>
  );
}
