"use client";

import * as React from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export interface PageHeaderProps {
  title: string;
  actions?: React.ReactNode;
}

export function PageHeader({ title, actions }: PageHeaderProps) {
  return (
    <div className="page-head">
      <div>
        <h1 data-testid="page-title">{title}</h1>
        {actions}
      </div>
    </div>
  );
}

export interface FilterBarProps {
  children: React.ReactNode;
}

export function FilterBar({ children }: FilterBarProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      {children}
    </div>
  );
}

export interface SurfaceCardProps {
  children: React.ReactNode;
  className?: string;
}

export function SurfaceCard({ children, className }: SurfaceCardProps) {
  return (
    <Card className={cn(className)}>
      {children}
    </Card>
  );
}

export interface DrawerPanelProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}

export function DrawerPanel({ open, onClose, title, children }: DrawerPanelProps) {
  if (!open) return null;
  return (
    <aside className="fixed inset-y-0 right-0 z-50 flex w-96 flex-col bg-background shadow-lg">
      <div className="flex items-center justify-between border-b p-4">
        <span className="font-medium">{title}</span>
        <Button variant="ghost" size="icon" data-testid="drawer-close" onClick={onClose}>
          ×
        </Button>
      </div>
      <div className="flex-1 overflow-y-auto p-4">
        {children}
      </div>
    </aside>
  );
}
