"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import type { NavItem } from "./sidebar-config";

interface SidebarProps {
  items: NavItem[];
  open?: boolean;
  onClose?: () => void;
}

export function Sidebar({ items, open = true, onClose }: SidebarProps) {
  const pathname = usePathname();

  return (
    <aside
      className={cn(
        "fixed inset-y-0 left-0 z-40 flex w-64 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-transform duration-200 ease-in-out md:translate-x-0",
        open ? "translate-x-0" : "-translate-x-full"
      )}
      aria-label="Main navigation"
    >
      <div className="flex h-14 shrink-0 items-center gap-2 border-b border-sidebar-border px-4">
        <img
          src="/logo.svg"
          alt="LLMGopher"
          className="h-8 w-8 shrink-0 object-contain"
        />
        <span className="text-base font-semibold tracking-tight">
          LLMGopher
        </span>
      </div>

      <nav className="flex-1 overflow-y-auto px-3 py-4" aria-label="Sidebar">
        <ul className="space-y-1">
          {items.map((item) => {
            const isActive =
              pathname === item.href ||
              (item.href !== "/" && pathname.startsWith(item.href));
            const Icon = item.icon;
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  onClick={onClose}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                    isActive
                      ? "border-l-4 border-l-primary bg-sidebar-accent text-sidebar-accent-foreground"
                      : "border-l-4 border-l-transparent text-sidebar-foreground/80 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                  )}
                  aria-current={isActive ? "page" : undefined}
                >
                  <Icon
                    className={cn(
                      "h-5 w-5 shrink-0",
                      isActive ? "text-primary" : "text-muted-foreground"
                    )}
                    aria-hidden
                  />
                  {item.name}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>
    </aside>
  );
}
