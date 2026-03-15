"use client";

import { useState } from "react";
import { Menu } from "lucide-react";
import { Sidebar } from "./Sidebar";
import { sidebarNavItems } from "./sidebar-config";
import { cn } from "@/lib/utils";

interface DashboardShellProps {
  children: React.ReactNode;
}

export function DashboardShell({ children }: DashboardShellProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="flex min-h-screen bg-muted/30">
      <Sidebar
        items={sidebarNavItems}
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
      />

      {/* Mobile overlay when sidebar is open */}
      {sidebarOpen && (
        <button
          type="button"
          aria-label="Close sidebar"
          className="fixed inset-0 z-30 bg-black/50 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      <div className="flex flex-1 flex-col min-w-0 md:pl-64">
        <header className="sticky top-0 z-20 flex h-14 shrink-0 items-center gap-4 border-b border-border bg-background px-4 md:px-6">
          <button
            type="button"
            aria-label="Open sidebar"
            onClick={() => setSidebarOpen(true)}
            className={cn(
              "rounded-lg p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground md:hidden"
            )}
          >
            <Menu className="h-5 w-5" aria-hidden />
          </button>
          <div className="flex-1" />
          {/* Placeholder for user profile / breadcrumbs */}
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            Control Plane
          </div>
        </header>

        <main className="flex-1 overflow-auto">
          <div className="p-6 md:p-8">{children}</div>
        </main>
      </div>
    </div>
  );
}
