import { deriveBudgetStatus, formatCurrencyUSD } from "@/lib/budget";
import { APIKeyBudgetState } from "@/lib/types";

type APIKeyBudgetStatusProps = {
  state: APIKeyBudgetState;
};

function formatResetAt(value?: string | null): string {
  if (!value) {
    return "Not scheduled";
  }
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? "Unknown" : parsed.toLocaleString();
}

export default function APIKeyBudgetStatus({ state }: APIKeyBudgetStatusProps) {
  if (state.status === "unavailable") {
    return <span className="text-xs text-muted-foreground">{state.message}</span>;
  }

  if (state.status === "unbudgeted") {
    return <span className="text-xs text-muted-foreground">No budget set</span>;
  }

  const status = deriveBudgetStatus(state);
  const tone =
    status === "exhausted"
      ? "text-destructive"
      : status === "near_threshold"
        ? "text-amber-600"
        : "text-emerald-600";

  return (
    <div className="space-y-1 text-xs">
      <div className={tone}>
        {status === "exhausted"
          ? "Exhausted"
          : status === "near_threshold"
            ? "Near threshold"
            : "Within budget"}
      </div>
      <div>
        {formatCurrencyUSD(state.budget.spent_usd)} spent /{" "}
        {formatCurrencyUSD(state.budget.budget_usd)}
      </div>
      <div>{formatCurrencyUSD(state.budget.remaining_usd)} remaining</div>
      {state.budget.alert_threshold_pct ? (
        <div>Alert at {state.budget.alert_threshold_pct}%</div>
      ) : null}
      {state.budget.budget_duration ? (
        <div>
          {state.budget.budget_duration} reset {formatResetAt(state.budget.budget_reset_at)}
        </div>
      ) : null}
    </div>
  );
}
