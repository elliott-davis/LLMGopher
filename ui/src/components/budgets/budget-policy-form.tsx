'use client';

import type { BudgetPolicy } from '@/lib/admin-surface-contracts';
import { formatCurrency, formatPercentage } from '@/lib/admin-surface-format';

export function BudgetPolicyForm({ budget }: { budget: BudgetPolicy }) {
  const usageNum = budget.usage_usd;
  const limitNum = budget.limit_usd;
  const utilization = limitNum > 0 ? usageNum / limitNum : 0;
  const isNearCap = budget.hard_cap_state === 'near_cap';
  const isOverCap = budget.hard_cap_state === 'over_cap';

  return (
    <div
      data-testid={`team-${budget.scope_id.replace(/_/g, '-')}-budget`}
      className={`border rounded p-4 ${isNearCap ? 'border-amber-300 bg-amber-50' : isOverCap ? 'border-red-300 bg-red-50' : ''}`}
    >
      {isNearCap && (
        <div
          data-testid={`team-${budget.scope_id.replace(/_/g, '-')}-warn`}
          className="text-xs text-amber-700 bg-amber-100 px-2 py-0.5 rounded inline-block mb-2"
        >
          Near cap — {formatPercentage(utilization)} used
        </div>
      )}
      {isOverCap && (
        <div className="text-xs text-red-700 bg-red-100 px-2 py-0.5 rounded inline-block mb-2">
          Over cap — {formatPercentage(utilization)} used
        </div>
      )}
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium capitalize">{budget.scope}: {budget.scope_id}</p>
          <p className="text-xs text-muted-foreground">
            {formatCurrency(usageNum)} / {formatCurrency(limitNum)} · {budget.duration}
          </p>
        </div>
      </div>
    </div>
  );
}
