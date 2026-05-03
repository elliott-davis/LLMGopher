"use client";

import { FormEvent, useMemo, useState } from "react";

import { APIKeyBudget, BudgetDuration } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type APIKeyBudgetFormProps = {
  initialBudget?: APIKeyBudget;
  disabled?: boolean;
  onSubmit: (formData: FormData) => Promise<void> | void;
};

type FieldErrors = Partial<Record<"budget_usd" | "alert_threshold_pct" | "budget_duration" | "budget_reset_at", string>>;

function toDateTimeLocal(value?: string | null): string {
  if (!value) {
    return "";
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "";
  }
  return parsed.toISOString().slice(0, 16);
}

export default function APIKeyBudgetForm({ initialBudget, disabled = false, onSubmit }: APIKeyBudgetFormProps) {
  const [budgetUSD, setBudgetUSD] = useState(String(initialBudget?.budget_usd ?? ""));
  const [alertThreshold, setAlertThreshold] = useState(
    initialBudget?.alert_threshold_pct ? String(initialBudget.alert_threshold_pct) : ""
  );
  const [duration, setDuration] = useState<BudgetDuration | "">(
    (initialBudget?.budget_duration as BudgetDuration | undefined) ?? ""
  );
  const [resetAt, setResetAt] = useState(toDateTimeLocal(initialBudget?.budget_reset_at));
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  const isDisabled = disabled || submitting;

  const validationErrors = useMemo(() => {
    const errors: FieldErrors = {};
    const budget = Number(budgetUSD);
    if (!Number.isFinite(budget) || budget <= 0) {
      errors.budget_usd = "Enter a budget limit greater than zero.";
    }
    if (alertThreshold) {
      const threshold = Number(alertThreshold);
      if (!Number.isInteger(threshold) || threshold < 1 || threshold > 99) {
        errors.alert_threshold_pct = "Alert threshold must be 1-99.";
      }
    }
    if (duration && !["daily", "weekly", "monthly"].includes(duration)) {
      errors.budget_duration = "Choose a valid duration.";
    }
    if (duration && !resetAt) {
      errors.budget_reset_at = "Reset time is required when duration is set.";
    }
    return errors;
  }, [alertThreshold, budgetUSD, duration, resetAt]);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setFieldErrors(validationErrors);
    if (Object.keys(validationErrors).length > 0) {
      return;
    }

    const formData = new FormData();
    formData.set("budget_usd", budgetUSD);
    formData.set("alert_threshold_pct", alertThreshold);
    formData.set("budget_duration", duration);
    formData.set("budget_reset_at", resetAt);

    setSubmitting(true);
    try {
      await onSubmit(formData);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form className="space-y-3" onSubmit={handleSubmit} noValidate>
      <div className="space-y-1">
        <Label htmlFor="budget-usd">Budget limit (USD)</Label>
        <Input
          id="budget-usd"
          name="budget_usd"
          type="number"
          min={0.01}
          step="0.01"
          value={budgetUSD}
          onChange={(event) => setBudgetUSD(event.target.value)}
          disabled={isDisabled}
        />
        {fieldErrors.budget_usd ? <p className="text-xs text-destructive">{fieldErrors.budget_usd}</p> : null}
      </div>

      <div className="space-y-1">
        <Label htmlFor="budget-threshold">Alert threshold (%)</Label>
        <Input
          id="budget-threshold"
          name="alert_threshold_pct"
          type="number"
          min={1}
          max={99}
          step={1}
          value={alertThreshold}
          onChange={(event) => setAlertThreshold(event.target.value)}
          disabled={isDisabled}
        />
        {fieldErrors.alert_threshold_pct ? (
          <p className="text-xs text-destructive">{fieldErrors.alert_threshold_pct}</p>
        ) : null}
      </div>

      <div className="space-y-1">
        <Label htmlFor="budget-duration">Budget duration</Label>
        <select
          id="budget-duration"
          name="budget_duration"
          className="h-9 w-full rounded-md border bg-background px-3 text-sm"
          value={duration}
          onChange={(event) => setDuration(event.target.value as BudgetDuration | "")}
          disabled={isDisabled}
        >
          <option value="">No auto reset</option>
          <option value="daily">Daily</option>
          <option value="weekly">Weekly</option>
          <option value="monthly">Monthly</option>
        </select>
        {fieldErrors.budget_duration ? (
          <p className="text-xs text-destructive">{fieldErrors.budget_duration}</p>
        ) : null}
      </div>

      <div className="space-y-1">
        <Label htmlFor="budget-reset-at">Reset time</Label>
        <Input
          id="budget-reset-at"
          name="budget_reset_at"
          type="datetime-local"
          value={resetAt}
          onChange={(event) => setResetAt(event.target.value)}
          disabled={isDisabled}
        />
        {fieldErrors.budget_reset_at ? (
          <p className="text-xs text-destructive">{fieldErrors.budget_reset_at}</p>
        ) : null}
      </div>

      <Button type="submit" disabled={isDisabled}>
        {submitting ? "Saving..." : "Save budget"}
      </Button>
    </form>
  );
}
