import type { BudgetPolicy } from '@/lib/admin-surface-contracts';
import { utilizationState } from '@/lib/admin-surface-format';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

interface RawBudget {
  scope: 'team' | 'key';
  scope_id: string;
  limit_usd: string;
  usage_usd: string;
  duration: 'daily' | 'weekly' | 'monthly';
  alert_threshold: number;
}

export async function fetchBudgets(): Promise<{ budgets: BudgetPolicy[]; unavailable: boolean }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/budgets`, { cache: 'no-store' });
    if (!res.ok) return { budgets: [], unavailable: true };
    const body = await res.json() as { data: RawBudget[] };
    const budgets: BudgetPolicy[] = (body.data ?? []).map((b) => {
      const limit = parseFloat(b.limit_usd);
      const usage = parseFloat(b.usage_usd);
      const utilization = limit > 0 ? usage / limit : 0;
      return {
        scope: b.scope,
        scope_id: b.scope_id,
        limit_usd: limit,
        usage_usd: usage,
        duration: b.duration,
        alert_threshold: b.alert_threshold,
        hard_cap_state: utilizationState(utilization, b.alert_threshold),
      };
    });
    return { budgets, unavailable: false };
  } catch {
    return { budgets: [], unavailable: true };
  }
}
