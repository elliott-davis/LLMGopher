import type { MockGuardrail } from "../mock/types";

export const guardrails: MockGuardrail[] = [
  { id: "gr_jail",    display_name: "Jailbreak detection", enabled: false },
  { id: "gr_pii",     display_name: "PII redaction",        enabled: true  },
  { id: "gr_secrets", display_name: "Secret scanning",      enabled: false },
];
