'use client';

import type { Team } from '@/lib/admin-surface-contracts';
import { PageHeader } from '@/components/admin-surface-layout';
import { UnavailableState, EmptyState } from '@/components/admin-surface-state';
import { TeamsList } from './teams-list';

export function TeamsClient({ teams, unavailable }: { teams: Team[]; unavailable: boolean }) {
  return (
    <>
      <PageHeader title="Teams" />
      <div className="content-area">
        {unavailable && <UnavailableState message="Teams are not yet available in this environment." />}
        {!unavailable && teams.length === 0 && <EmptyState message="No teams configured." />}
        {!unavailable && teams.length > 0 && <TeamsList teams={teams} />}
      </div>
    </>
  );
}
