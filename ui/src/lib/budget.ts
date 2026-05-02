import {
  APIKeyBudget,
  APIKeyBudgetFormValues,
  APIKeyBudgetState,
  BudgetDuration,
  BudgetStatusIndicator,
} from "@/lib/types";

const VALID_DURATIONS: BudgetDuration[] = ["daily", "weekly", "monthly"];

function parseOptionalNumber(value: unknown): number | null {
  if (value === null || value === undefined || value === "") {
    return null;
  }
  const num = Number(value);
  return Number.isFinite(num) ? num : null;
}

function parseOptionalInteger(value: unknown): number | null {
  const num = parseOptionalNumber(value);
  if (num === null || !Number.isInteger(num)) {
    return null;
  }
  return num;
}

export function parseAPIKeyBudget(payload: unknown): APIKeyBudget {
  if (typeof payload !== "object" || payload === null) {
    throw new Error("Invalid budget response");
  }

  const raw = payload as Record<string, unknown>;
  const apiKeyID = String(raw.api_key_id ?? "").trim();
  const budgetUSD = parseOptionalNumber(raw.budget_usd);
  const spentUSD = parseOptionalNumber(raw.spent_usd);
  const remainingUSD = parseOptionalNumber(raw.remaining_usd);
  const threshold = parseOptionalInteger(raw.alert_threshold_pct);
  const durationRaw = raw.budget_duration;
  const resetAtRaw = raw.budget_reset_at;

  if (!apiKeyID || budgetUSD === null || spentUSD === null || remainingUSD === null || budgetUSD <= 0) {
    throw new Error("Invalid budget response");
  }
  if (spentUSD < 0) {
    throw new Error("Invalid budget response");
  }
  if (threshold !== null && (threshold < 1 || threshold > 99)) {
    throw new Error("Invalid budget response");
  }

  let duration: BudgetDuration | undefined;
  if (durationRaw !== null && durationRaw !== undefined && durationRaw !== "") {
    if (typeof durationRaw !== "string" || !VALID_DURATIONS.includes(durationRaw as BudgetDuration)) {
      throw new Error("Invalid budget response");
    }
    duration = durationRaw as BudgetDuration;
  }

  let resetAt: string | undefined;
  if (resetAtRaw !== null && resetAtRaw !== undefined && resetAtRaw !== "") {
    if (typeof resetAtRaw !== "string") {
      throw new Error("Invalid budget response");
    }
    const parsed = new Date(resetAtRaw);
    if (Number.isNaN(parsed.getTime())) {
      throw new Error("Invalid budget response");
    }
    resetAt = parsed.toISOString();
  }

  return {
    api_key_id: apiKeyID,
    budget_usd: budgetUSD,
    spent_usd: spentUSD,
    remaining_usd: remainingUSD,
    alert_threshold_pct: threshold ?? undefined,
    budget_duration: duration,
    budget_reset_at: resetAt,
  };
}

export function deriveBudgetStatus(state: APIKeyBudgetState): BudgetStatusIndicator {
  if (state.status === "unavailable") {
    return "unavailable";
  }
  if (state.status === "unbudgeted") {
    return "unbudgeted";
  }

  const { budget } = state;
  if (budget.remaining_usd <= 0) {
    return "exhausted";
  }

  if (budget.alert_threshold_pct && budget.budget_usd > 0) {
    const spentPct = (budget.spent_usd / budget.budget_usd) * 100;
    if (spentPct >= budget.alert_threshold_pct) {
      return "near_threshold";
    }
  }

  return "within_budget";
}

export function formatCurrencyUSD(value: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

export function parseAPIKeyBudgetFormValues(formData: FormData): APIKeyBudgetFormValues {
  const budgetText = String(formData.get("budget_usd") ?? "").trim();
  const thresholdText = String(formData.get("alert_threshold_pct") ?? "").trim();
  const durationText = String(formData.get("budget_duration") ?? "").trim();
  const resetText = String(formData.get("budget_reset_at") ?? "").trim();

  const budgetUSD = Number(budgetText);
  if (!Number.isFinite(budgetUSD) || budgetUSD <= 0) {
    throw new Error("Budget limit must be greater than zero");
  }

  let threshold: number | undefined;
  if (thresholdText) {
    const parsedThreshold = Number(thresholdText);
    if (!Number.isInteger(parsedThreshold) || parsedThreshold < 1 || parsedThreshold > 99) {
      throw new Error("Alert threshold must be an integer between 1 and 99");
    }
    threshold = parsedThreshold;
  }

  let duration: BudgetDuration | undefined;
  if (durationText) {
    if (!VALID_DURATIONS.includes(durationText as BudgetDuration)) {
      throw new Error("Budget duration must be daily, weekly, or monthly");
    }
    duration = durationText as BudgetDuration;
  }

  let resetAt: string | undefined;
  if (resetText) {
    const parsed = new Date(resetText);
    if (Number.isNaN(parsed.getTime())) {
      throw new Error("Budget reset time must be a valid date");
    }
    resetAt = parsed.toISOString();
  }

  if (duration && !resetAt) {
    throw new Error("Budget reset time is required when duration is set");
  }

  return {
    budget_usd: budgetUSD,
    alert_threshold_pct: threshold,
    budget_duration: duration,
    budget_reset_at: resetAt,
  };
}
