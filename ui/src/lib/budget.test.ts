import { describe, expect, it } from "vitest";

import {
  deriveBudgetStatus,
  parseAPIKeyBudget,
  parseAPIKeyBudgetFormValues,
} from "@/lib/budget";
import { APIKeyBudgetState } from "@/lib/types";

describe("budget helpers", () => {
  it("derives configured status variants", () => {
    const within: APIKeyBudgetState = {
      status: "configured",
      budget: {
        api_key_id: "key-1",
        budget_usd: 100,
        spent_usd: 20,
        remaining_usd: 80,
        alert_threshold_pct: 80,
      },
    };
    const near: APIKeyBudgetState = {
      status: "configured",
      budget: {
        api_key_id: "key-1",
        budget_usd: 100,
        spent_usd: 85,
        remaining_usd: 15,
        alert_threshold_pct: 80,
      },
    };
    const exhausted: APIKeyBudgetState = {
      status: "configured",
      budget: {
        api_key_id: "key-1",
        budget_usd: 100,
        spent_usd: 100,
        remaining_usd: 0,
      },
    };

    expect(deriveBudgetStatus(within)).toBe("within_budget");
    expect(deriveBudgetStatus(near)).toBe("near_threshold");
    expect(deriveBudgetStatus(exhausted)).toBe("exhausted");
    expect(deriveBudgetStatus({ status: "unbudgeted" })).toBe("unbudgeted");
    expect(deriveBudgetStatus({ status: "unavailable", message: "auth required" })).toBe(
      "unavailable"
    );
  });

  it("parses valid gateway budget responses", () => {
    expect(
      parseAPIKeyBudget({
        api_key_id: "key-1",
        budget_usd: 100,
        spent_usd: 25.5,
        remaining_usd: 74.5,
        alert_threshold_pct: 80,
        budget_duration: "monthly",
        budget_reset_at: "2026-06-01T00:00:00Z",
      })
    ).toMatchObject({
      api_key_id: "key-1",
      budget_usd: 100,
      spent_usd: 25.5,
      remaining_usd: 74.5,
      alert_threshold_pct: 80,
      budget_duration: "monthly",
    });
  });

  it("rejects invalid gateway values", () => {
    expect(() =>
      parseAPIKeyBudget({
        api_key_id: "key-1",
        budget_usd: 0,
        spent_usd: 1,
        remaining_usd: -1,
      })
    ).toThrow("Invalid budget response");

    expect(() =>
      parseAPIKeyBudget({
        api_key_id: "key-1",
        budget_usd: 10,
        spent_usd: 1,
        remaining_usd: 9,
        alert_threshold_pct: 120,
      })
    ).toThrow("Invalid budget response");
  });

  it("validates budget form rules", () => {
    const valid = new FormData();
    valid.set("budget_usd", "100");
    valid.set("alert_threshold_pct", "80");
    valid.set("budget_duration", "monthly");
    valid.set("budget_reset_at", "2026-06-01T00:00");
    expect(parseAPIKeyBudgetFormValues(valid)).toMatchObject({
      budget_usd: 100,
      alert_threshold_pct: 80,
      budget_duration: "monthly",
    });

    const invalidLimit = new FormData();
    invalidLimit.set("budget_usd", "0");
    expect(() => parseAPIKeyBudgetFormValues(invalidLimit)).toThrow(
      "Budget limit must be greater than zero"
    );

    const invalidThreshold = new FormData();
    invalidThreshold.set("budget_usd", "10");
    invalidThreshold.set("alert_threshold_pct", "100");
    expect(() => parseAPIKeyBudgetFormValues(invalidThreshold)).toThrow(
      "Alert threshold must be an integer between 1 and 99"
    );

    const missingReset = new FormData();
    missingReset.set("budget_usd", "10");
    missingReset.set("budget_duration", "monthly");
    expect(() => parseAPIKeyBudgetFormValues(missingReset)).toThrow(
      "Budget reset time is required when duration is set"
    );
  });
});
