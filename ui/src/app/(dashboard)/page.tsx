import Link from "next/link";
import { KeyRound, GitBranch, Box } from "lucide-react";

const cards = [
  {
    title: "API Keys",
    href: "/keys",
    description: "View and create gateway API keys used by clients.",
    icon: KeyRound,
  },
  {
    title: "Models & Routing",
    href: "/models",
    description:
      "Review configured providers, model aliases, and future routing policies.",
    icon: GitBranch,
  },
  {
    title: "Providers",
    href: "/providers",
    description:
      "Create provider instances and use them from the Add Model workflow.",
    icon: Box,
  },
] as const;

export default function DashboardPage() {
  return (
    <div className="mx-auto w-full max-w-7xl">
      <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        Manage gateway access, inspect model routing, and operate the local
        development control plane.
      </p>
      <div className="mt-8 grid gap-6 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {cards.map(({ title, href, description, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className="group flex flex-col rounded-xl border border-gray-200 bg-white p-6 text-left shadow-sm transition-all hover:shadow-md focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
          >
            <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-lg bg-muted text-muted-foreground transition-colors group-hover:bg-primary/10 group-hover:text-primary">
              <Icon className="h-5 w-5" aria-hidden />
            </div>
            <h2 className="text-base font-semibold tracking-tight text-foreground">
              {title}
            </h2>
            <p className="mt-1 flex-1 text-sm text-muted-foreground">
              {description}
            </p>
            <span className="mt-3 text-sm font-medium text-primary group-hover:underline">
              Open →
            </span>
          </Link>
        ))}
      </div>
    </div>
  );
}
