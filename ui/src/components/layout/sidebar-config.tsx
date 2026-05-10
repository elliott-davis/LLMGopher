export interface NavItem {
  name: string;
  href: string;
  iconPath: string;
  badge?: string; // numeric string or "live"
  testId: string; // stable id for E2E selectors (data-testid="nav-{testId}")
}

export interface NavGroup {
  group: string;
  items: NavItem[];
}

export const sidebarNav: NavGroup[] = [
  {
    group: "Operate",
    items: [
      {
        name: "Overview",
        href: "/",
        iconPath: "M3 13h8V3H3zM13 21h8V11h-8zM3 21h8v-6H3zM13 9h8V3h-8z",
        testId: "overview",
      },
      {
        name: "Logs",
        href: "/logs",
        iconPath:
          "M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M9 13h6 M9 17h6",
        badge: "live",
        testId: "logs",
      },
      {
        name: "Audit log",
        href: "/audit",
        iconPath:
          "M9 11l3 3L22 4 M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11",
        testId: "audit",
      },
    ],
  },
  {
    group: "Configure",
    items: [
      {
        name: "Providers",
        href: "/providers",
        iconPath: "M4 7h16M4 12h16M4 17h16 M8 7v10 M16 7v10",
        testId: "providers",
      },
      {
        name: "Routes",
        href: "/routes",
        iconPath:
          "M3 6h6a4 4 0 0 1 4 4v4a4 4 0 0 0 4 4h4 M17 14l4 4-4 4",
        testId: "routes",
      },
      {
        name: "Guardrails",
        href: "/guardrails",
        iconPath:
          "M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z",
        testId: "guardrails",
      },
    ],
  },
  {
    group: "Govern",
    items: [
      {
        name: "API Keys",
        href: "/keys",
        iconPath:
          "M21 2 13 10 m-2 2a3 3 0 1 1-4 4 3 3 0 0 1 4-4z m4-6 3 3 m-5-1 3 3",
        testId: "keys",
      },
      {
        name: "Teams",
        href: "/teams",
        iconPath:
          "M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2 M22 21v-2a4 4 0 0 0-3-3.87 M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z M16 3.13a4 4 0 0 1 0 7.75",
        testId: "teams",
      },
      {
        name: "Budgets",
        href: "/budgets",
        iconPath:
          "M12 2v20 M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6",
        testId: "budgets",
      },
      {
        name: "Rate limits",
        href: "/rate-limits",
        iconPath:
          "M12 8v4l3 2 M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z",
        testId: "rate",
      },
    ],
  },
  {
    group: "Org",
    items: [
      {
        name: "Settings",
        href: "/settings",
        testId: "settings",
        iconPath:
          "M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33h0a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51h0a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82v0a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z",
      },
    ],
  },
];

export function getBreadcrumb(pathname: string): [string, string] {
  for (const group of sidebarNav) {
    for (const item of group.items) {
      if (
        item.href === pathname ||
        (item.href !== "/" && pathname.startsWith(item.href))
      ) {
        return [group.group, item.name];
      }
    }
  }
  return ["", ""];
}
