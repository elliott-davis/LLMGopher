import type { APIKey } from "../../src/lib/types";

const NOW = "2026-05-09T12:00:00.000Z";

export interface KeyFixture extends APIKey {
  // Budget state carried for mock handlers — not in the base APIKey type.
  _budget?: {
    limit_usd: number;
    usage_usd: number;
    duration: "daily" | "weekly" | "monthly";
    alert_threshold: number;
  };
}

export const keys: KeyFixture[] = [
  {
    id: "key_checkout_service",
    key_hash: "sha256:fake0001",
    name: "checkout-service",
    rate_limit_rps: 100,
    is_active: true,
    expires_at: null,
    allowed_models: null,
    created_at: NOW,
    updated_at: NOW,
  },
  {
    id: "key_research_team",
    key_hash: "sha256:fake0002",
    name: "research-team",
    rate_limit_rps: 50,
    is_active: true,
    expires_at: null,
    allowed_models: ["gpt-4o"],
    created_at: NOW,
    updated_at: NOW,
  },
  {
    id: "key_near_cap",
    key_hash: "sha256:fake0003",
    name: "near-cap-service",
    rate_limit_rps: 20,
    is_active: true,
    expires_at: null,
    allowed_models: null,
    created_at: NOW,
    updated_at: NOW,
    // 86% utilization → triggers "near cap" warning pill
    _budget: { limit_usd: 1000, usage_usd: 860, duration: "monthly", alert_threshold: 0.85 },
  },
  {
    id: "key_over_cap",
    key_hash: "sha256:fake0004",
    name: "over-cap-service",
    rate_limit_rps: 20,
    is_active: true,
    expires_at: null,
    allowed_models: null,
    created_at: NOW,
    updated_at: NOW,
    // 100% utilization → triggers 429 budget_exceeded on chat completions
    _budget: { limit_usd: 500, usage_usd: 500, duration: "monthly", alert_threshold: 0.85 },
  },
];
