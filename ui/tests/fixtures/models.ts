import type { Model } from "../../src/lib/types";

const NOW = "2026-05-09T12:00:00.000Z";

export const models: Model[] = [
  { id: "gpt-4o",            provider_id: "prov_openai_prod",        name: "GPT-4o",               alias: "gpt-4o",            context_window: 128000, rate_limit_rps: 100, created_at: NOW, updated_at: NOW },
  { id: "gpt-4o-mini",       provider_id: "prov_openai_prod",        name: "GPT-4o Mini",          alias: "gpt-4o-mini",       context_window: 128000, rate_limit_rps: 200, created_at: NOW, updated_at: NOW },
  { id: "gpt-3.5-turbo",     provider_id: "prov_openai_prod",        name: "GPT-3.5 Turbo",        alias: "gpt-3.5-turbo",     context_window: 16385,  rate_limit_rps: 0,   created_at: NOW, updated_at: NOW },
  { id: "claude-3-5-sonnet", provider_id: "prov_anthropic_degraded", name: "Claude 3.5 Sonnet",    alias: "claude-3.5-sonnet", context_window: 200000, rate_limit_rps: 50,  created_at: NOW, updated_at: NOW },
  { id: "claude-3-haiku",    provider_id: "prov_anthropic_degraded", name: "Claude 3 Haiku",       alias: "claude-3-haiku",    context_window: 200000, rate_limit_rps: 100, created_at: NOW, updated_at: NOW },
  { id: "gemini-1-5-pro",    provider_id: "prov_vertex_offline",     name: "Gemini 1.5 Pro",       alias: "gemini-1.5-pro",    context_window: 1000000, rate_limit_rps: 0,  created_at: NOW, updated_at: NOW },
];
