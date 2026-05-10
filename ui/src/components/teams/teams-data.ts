import type { Team } from '@/lib/admin-surface-contracts';
import { utilizationState } from '@/lib/admin-surface-format';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function fetchTeams(): Promise<{ teams: Team[]; unavailable: boolean }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/teams`, { cache: 'no-store' });
    if (!res.ok) return { teams: [], unavailable: true };
    const body = await res.json() as { data: Array<{ id: string; display_name: string; member_count: number; budget_utilization: number }> };
    const teams: Team[] = (body.data ?? []).map((t) => ({
      ...t,
      budget_health: utilizationState(t.budget_utilization, 0.85),
    }));
    return { teams, unavailable: false };
  } catch {
    return { teams: [], unavailable: true };
  }
}
