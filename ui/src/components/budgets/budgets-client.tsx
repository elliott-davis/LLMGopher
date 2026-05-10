'use client';

import type { BudgetPolicy } from '@/lib/admin-surface-contracts';
import { PageHeader } from '@/components/admin-surface-layout';
import { UnavailableState, EmptyState } from '@/components/admin-surface-state';
import { BudgetPolicyForm } from './budget-policy-form';

export function BudgetsClient({ budgets, unavailable }: { budgets: BudgetPolicy[]; unavailable: boolean }) {
  return (
    <>
      <PageHeader title="Budgets" />
      <div className="content-area">
        {unavailable && <UnavailableState message="Budgets are not yet available in this environment." />}
        {!unavailable && budgets.length === 0 && <EmptyState message="No budget policies configured." />}
        {!unavailable && budgets.length > 0 && (
          <div className="space-y-3">
            {budgets.map((budget) => (
              <BudgetPolicyForm key={`${budget.scope}-${budget.scope_id}`} budget={budget} />
            ))}
          </div>
        )}
      </div>
    </>
  );
}
