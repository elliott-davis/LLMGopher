'use client';

import type { Team } from '@/lib/admin-surface-contracts';
import { formatPercentage } from '@/lib/admin-surface-format';

export function TeamsList({ teams }: { teams: Team[] }) {
  return (
    <div className="space-y-3">
      {teams.map((team) => (
        <div
          key={team.id}
          data-testid={`team-row-${team.id}`}
          className="border rounded p-4 flex items-center justify-between"
        >
          <div>
            <p className="font-medium text-sm">{team.display_name}</p>
            <p className="text-xs text-muted-foreground">{team.member_count} members</p>
          </div>
          <div className="text-right">
            <p className="text-sm font-medium">{formatPercentage(team.budget_utilization)}</p>
            {team.budget_health === 'near_cap' && (
              <span
                data-testid={`team-${team.id.replace(/_/g, '-')}-warn`}
                className="text-xs text-amber-700 bg-amber-100 px-2 py-0.5 rounded"
                aria-label={`${team.display_name} budget near cap`}
              >
                Near cap
              </span>
            )}
            {team.budget_health === 'over_cap' && (
              <span className="text-xs text-red-700 bg-red-100 px-2 py-0.5 rounded">Over cap</span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
