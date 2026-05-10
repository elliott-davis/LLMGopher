import type { MockRateLimitRule } from "../mock/types";

export const rateLimits: MockRateLimitRule[] = [
  { id: "rl_chat_default", scope: "model",  scope_id: "gpt-4o",               rps: 100, tpm: 60000, tripped: false },
  { id: "rl_key_checkout",  scope: "key",   scope_id: "key_checkout_service",  rps: 50,              tripped: false },
  // Exactly one rule is tripped so the "tripped" pill test has a deterministic target.
  { id: "rl_tripped",       scope: "model", scope_id: "claude-3.5-sonnet",     rps: 10,              tripped: true  },
];
