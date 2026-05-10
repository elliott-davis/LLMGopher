import type { MockLogRow } from "../mock/types";

const BASE = 1762689600000; // 2026-05-09T12:00:00.000Z

function isoOffset(ms: number): string {
  return new Date(BASE - ms).toISOString();
}

// 20 log rows with mixed status. Row log_fallback has a multi-stage provider_chain
// where the first stage failed (used by the trace-tab test).
export const logs: MockLogRow[] = [
  { id: "log_001", request_id: "req_0001", timestamp: isoOffset(10_000),  method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 432, api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 410 }] },
  { id: "log_002", request_id: "req_0002", timestamp: isoOffset(20_000),  method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 289, api_key_id: "key_research_team",    model: "claude-3.5-sonnet", provider_chain: [{ provider_id: "prov_anthropic_degraded", status: "ok",     latency_ms: 270 }] },
  { id: "log_003", request_id: "req_0003", timestamp: isoOffset(30_000),  method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 512, api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 495 }] },
  { id: "log_004", request_id: "req_0004", timestamp: isoOffset(40_000),  method: "POST", path: "/v1/chat/completions", status_code: 429, latency_ms: 5,   api_key_id: "key_over_cap",         model: "gpt-4o-mini",       provider_chain: [] },
  { id: "log_005", request_id: "req_0005", timestamp: isoOffset(50_000),  method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 654, api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 632 }] },
  { id: "log_006", request_id: "req_0006", timestamp: isoOffset(60_000),  method: "POST", path: "/v1/chat/completions", status_code: 500, latency_ms: 120, api_key_id: "key_research_team",    model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "failed", latency_ms: 100 }] },
  { id: "log_007", request_id: "req_0007", timestamp: isoOffset(70_000),  method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 388, api_key_id: "key_checkout_service", model: "gpt-4o-mini",       provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 370 }] },
  { id: "log_008", request_id: "req_0008", timestamp: isoOffset(80_000),  method: "POST", path: "/v1/chat/completions", status_code: 401, latency_ms: 8,   api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [] },
  { id: "log_009", request_id: "req_0009", timestamp: isoOffset(90_000),  method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 722, api_key_id: "key_research_team",    model: "claude-3.5-sonnet", provider_chain: [{ provider_id: "prov_anthropic_degraded", status: "ok",     latency_ms: 700 }] },
  { id: "log_010", request_id: "req_0010", timestamp: isoOffset(100_000), method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 445, api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 430 }] },
  { id: "log_011", request_id: "req_0011", timestamp: isoOffset(110_000), method: "POST", path: "/v1/chat/completions", status_code: 503, latency_ms: 30,  api_key_id: "key_research_team",    model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "failed", latency_ms: 25  }] },
  { id: "log_012", request_id: "req_0012", timestamp: isoOffset(120_000), method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 301, api_key_id: "key_checkout_service", model: "gpt-4o-mini",       provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 284 }] },
  { id: "log_013", request_id: "req_0013", timestamp: isoOffset(130_000), method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 589, api_key_id: "key_research_team",    model: "claude-3-haiku",    provider_chain: [{ provider_id: "prov_anthropic_degraded", status: "ok",     latency_ms: 568 }] },
  { id: "log_014", request_id: "req_0014", timestamp: isoOffset(140_000), method: "POST", path: "/v1/chat/completions", status_code: 422, latency_ms: 11,  api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [] },
  { id: "log_015", request_id: "req_0015", timestamp: isoOffset(150_000), method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 478, api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 460 }] },
  { id: "log_016", request_id: "req_0016", timestamp: isoOffset(160_000), method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 555, api_key_id: "key_research_team",    model: "claude-3.5-sonnet", provider_chain: [{ provider_id: "prov_anthropic_degraded", status: "ok",     latency_ms: 535 }] },
  { id: "log_017", request_id: "req_0017", timestamp: isoOffset(170_000), method: "POST", path: "/v1/chat/completions", status_code: 502, latency_ms: 45,  api_key_id: "key_checkout_service", model: "gpt-4o",            provider_chain: [{ provider_id: "prov_openai_prod",       status: "failed", latency_ms: 40  }] },
  { id: "log_018", request_id: "req_0018", timestamp: isoOffset(180_000), method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 398, api_key_id: "key_checkout_service", model: "gpt-4o-mini",       provider_chain: [{ provider_id: "prov_openai_prod",       status: "ok",     latency_ms: 381 }] },
  { id: "log_019", request_id: "req_0019", timestamp: isoOffset(190_000), method: "POST", path: "/v1/chat/completions", status_code: 200, latency_ms: 267, api_key_id: "key_research_team",    model: "claude-3-haiku",    provider_chain: [{ provider_id: "prov_anthropic_degraded", status: "ok",     latency_ms: 251 }] },
  // log_fallback: primary provider failed, fell back to secondary — used by trace-tab test.
  {
    id: "log_fallback",
    request_id: "req_0020",
    timestamp: isoOffset(200_000),
    method: "POST",
    path: "/v1/chat/completions",
    status_code: 200,
    latency_ms: 845,
    api_key_id: "key_checkout_service",
    model: "gpt-4o",
    provider_chain: [
      { provider_id: "prov_openai_prod",       status: "failed", latency_ms: 120 },
      { provider_id: "prov_anthropic_degraded", status: "ok",     latency_ms: 705 },
    ],
  },
];
