import type { AuditRecord } from "../../src/lib/types";

const BASE = 1762689600000; // 2026-05-09T12:00:00.000Z

function isoOffset(ms: number): string {
  return new Date(BASE - ms).toISOString();
}

// Matches the real AuditRecord interface from src/lib/types.ts.
export const auditEntries: AuditRecord[] = [
  { id: 1,  request_id: "req_a001", api_key_id: "key_checkout_service", model: "gpt-4o",            provider: "OpenAI · prod",    prompt_tokens: 120, output_tokens: 80,  total_tokens: 200, cost_usd: 0.0024, status_code: 200, latency_ms: 432, streaming: false, error_message: "", created_at: isoOffset(10_000)  },
  { id: 2,  request_id: "req_a002", api_key_id: "key_research_team",    model: "claude-3.5-sonnet", provider: "Anthropic · prod", prompt_tokens: 200, output_tokens: 150, total_tokens: 350, cost_usd: 0.0052, status_code: 200, latency_ms: 289, streaming: false, error_message: "", created_at: isoOffset(20_000)  },
  { id: 3,  request_id: "req_a003", api_key_id: "key_checkout_service", model: "gpt-4o",            provider: "OpenAI · prod",    prompt_tokens: 80,  output_tokens: 60,  total_tokens: 140, cost_usd: 0.0018, status_code: 200, latency_ms: 512, streaming: false, error_message: "", created_at: isoOffset(30_000)  },
  { id: 4,  request_id: "req_a004", api_key_id: "key_over_cap",         model: "gpt-4o-mini",       provider: "OpenAI · prod",    prompt_tokens: 0,   output_tokens: 0,   total_tokens: 0,   cost_usd: 0,      status_code: 429, latency_ms: 5,   streaming: false, error_message: "budget exceeded", created_at: isoOffset(40_000)  },
  { id: 5,  request_id: "req_a005", api_key_id: "key_checkout_service", model: "gpt-4o",            provider: "OpenAI · prod",    prompt_tokens: 300, output_tokens: 200, total_tokens: 500, cost_usd: 0.0060, status_code: 200, latency_ms: 654, streaming: false, error_message: "", created_at: isoOffset(50_000)  },
  { id: 6,  request_id: "req_a006", api_key_id: "key_research_team",    model: "gpt-4o",            provider: "OpenAI · prod",    prompt_tokens: 100, output_tokens: 0,   total_tokens: 100, cost_usd: 0.0010, status_code: 500, latency_ms: 120, streaming: false, error_message: "provider error",  created_at: isoOffset(60_000)  },
  { id: 7,  request_id: "req_a007", api_key_id: "key_checkout_service", model: "gpt-4o-mini",       provider: "OpenAI · prod",    prompt_tokens: 90,  output_tokens: 70,  total_tokens: 160, cost_usd: 0.0008, status_code: 200, latency_ms: 388, streaming: false, error_message: "", created_at: isoOffset(70_000)  },
  { id: 8,  request_id: "req_a008", api_key_id: "key_checkout_service", model: "gpt-4o",            provider: "OpenAI · prod",    prompt_tokens: 0,   output_tokens: 0,   total_tokens: 0,   cost_usd: 0,      status_code: 401, latency_ms: 8,   streaming: false, error_message: "unauthorized",    created_at: isoOffset(80_000)  },
  { id: 9,  request_id: "req_a009", api_key_id: "key_research_team",    model: "claude-3.5-sonnet", provider: "Anthropic · prod", prompt_tokens: 400, output_tokens: 300, total_tokens: 700, cost_usd: 0.0105, status_code: 200, latency_ms: 722, streaming: false, error_message: "", created_at: isoOffset(90_000)  },
  { id: 10, request_id: "req_a010", api_key_id: "key_checkout_service", model: "gpt-4o",            provider: "OpenAI · prod",    prompt_tokens: 150, output_tokens: 100, total_tokens: 250, cost_usd: 0.0030, status_code: 200, latency_ms: 445, streaming: false, error_message: "", created_at: isoOffset(100_000) },
];
