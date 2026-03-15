import type { LucideIcon } from "lucide-react";
import {
  KeyRound,
  LayoutDashboard,
  Box,
  GitBranch,
} from "lucide-react";

export interface NavItem {
  name: string;
  href: string;
  icon: LucideIcon;
}

export const sidebarNavItems: NavItem[] = [
  { name: "Dashboard", href: "/", icon: LayoutDashboard },
  { name: "API Keys", href: "/keys", icon: KeyRound },
  { name: "Models & Routing", href: "/models", icon: GitBranch },
  { name: "Providers", href: "/providers", icon: Box },
];
