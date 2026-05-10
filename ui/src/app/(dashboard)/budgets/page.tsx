import { fetchBudgets } from '@/components/budgets/budgets-data';
import { BudgetsClient } from '@/components/budgets/budgets-client';

export default async function BudgetsPage() {
  const { budgets, unavailable } = await fetchBudgets();
  return <BudgetsClient budgets={budgets} unavailable={unavailable} />;
}
