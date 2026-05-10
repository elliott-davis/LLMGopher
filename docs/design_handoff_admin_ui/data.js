// Mock data for LLMGopher Admin
window.LG_DATA = (function () {
  const providers = [
    { id: "prv_openai_main", name: "OpenAI · prod", kind: "openai", logo: "openai", base: "https://api.openai.com/v1", region: "us-east", status: "healthy", latencyP50: 412, latencyP95: 980, errors24h: 12, deployments: 3 },
    { id: "prv_anthropic_main", name: "Anthropic · prod", kind: "anthropic", logo: "anthropic", base: "https://api.anthropic.com", region: "us-west", status: "healthy", latencyP50: 590, latencyP95: 1240, errors24h: 4, deployments: 2 },
    { id: "prv_vertex_main", name: "Vertex · gemini", kind: "google", logo: "google", base: "vertexai://us-central1", region: "us-central", status: "degraded", latencyP50: 880, latencyP95: 2450, errors24h: 41, deployments: 2 },
    { id: "prv_mistral_eu", name: "Mistral · eu", kind: "mistral", logo: "mistral", base: "https://api.mistral.ai", region: "eu-west", status: "healthy", latencyP50: 320, latencyP95: 720, errors24h: 1, deployments: 1 },
    { id: "prv_bedrock_us", name: "Bedrock · us-east", kind: "bedrock", logo: "bedrock", base: "bedrock://us-east-1", region: "us-east", status: "healthy", latencyP50: 540, latencyP95: 1110, errors24h: 7, deployments: 4 },
    { id: "prv_vllm_local", name: "vLLM · staging-gpu", kind: "vllm", logo: "vllm", base: "http://10.0.4.12:8000", region: "self-hosted", status: "offline", latencyP50: 0, latencyP95: 0, errors24h: 0, deployments: 1 },
  ];

  const models = [
    { id: "gpt-4o", display: "gpt-4o", provider: "prv_openai_main", logo: "openai", ctx: 128000, in: 2.5, out: 10.0, rpm: 10000, tpm: 800000, route: "chat-prod" },
    { id: "gpt-4o-mini", display: "gpt-4o-mini", provider: "prv_openai_main", logo: "openai", ctx: 128000, in: 0.15, out: 0.6, rpm: 30000, tpm: 2000000, route: "chat-cheap" },
    { id: "claude-sonnet-4-5", display: "claude-sonnet-4.5", provider: "prv_anthropic_main", logo: "anthropic", ctx: 200000, in: 3.0, out: 15.0, rpm: 4000, tpm: 400000, route: "chat-prod" },
    { id: "claude-haiku-4-5", display: "claude-haiku-4.5", provider: "prv_anthropic_main", logo: "anthropic", ctx: 200000, in: 0.25, out: 1.25, rpm: 8000, tpm: 1000000, route: "chat-cheap" },
    { id: "vertex/gemini-2.5-pro", display: "gemini-2.5-pro", provider: "prv_vertex_main", logo: "google", ctx: 1000000, in: 1.25, out: 5.0, rpm: 1500, tpm: 300000, route: "chat-long" },
    { id: "vertex/gemini-2.5-flash", display: "gemini-2.5-flash", provider: "prv_vertex_main", logo: "google", ctx: 1000000, in: 0.075, out: 0.3, rpm: 6000, tpm: 1000000, route: "chat-cheap" },
    { id: "mistral-large", display: "mistral-large-2", provider: "prv_mistral_eu", logo: "mistral", ctx: 128000, in: 2.0, out: 6.0, rpm: 2000, tpm: 200000, route: "chat-eu" },
    { id: "bedrock/llama-3.3-70b", display: "llama-3.3-70b", provider: "prv_bedrock_us", logo: "bedrock", ctx: 128000, in: 0.72, out: 0.72, rpm: 1000, tpm: 100000, route: "embeddings" },
    { id: "vllm/llama-3.1-8b", display: "llama-3.1-8b", provider: "prv_vllm_local", logo: "vllm", ctx: 32000, in: 0, out: 0, rpm: 0, tpm: 0, route: "internal" },
  ];

  const routes = [
    {
      id: "chat-prod", name: "chat-prod",
      strategy: "fallback",
      desc: "Primary chat traffic. Sonnet → GPT-4o on failure.",
      members: [
        { model: "claude-sonnet-4-5", logo: "anthropic", weight: 100, role: "primary" },
        { model: "gpt-4o", logo: "openai", weight: 100, role: "fallback" },
      ],
      rpm: 18432, errRate: 0.4, p95: 1180, attached: 14
    },
    {
      id: "chat-cheap", name: "chat-cheap",
      strategy: "weighted",
      desc: "Cost-optimized chat. 60/30/10 between cheap models.",
      members: [
        { model: "gpt-4o-mini", logo: "openai", weight: 60, role: "active" },
        { model: "claude-haiku-4-5", logo: "anthropic", weight: 30, role: "active" },
        { model: "vertex/gemini-2.5-flash", logo: "google", weight: 10, role: "active" },
      ],
      rpm: 42118, errRate: 0.1, p95: 720, attached: 9
    },
    {
      id: "chat-long", name: "chat-long",
      strategy: "single",
      desc: "Long-context (>200k tokens) traffic.",
      members: [
        { model: "vertex/gemini-2.5-pro", logo: "google", weight: 100, role: "active" },
      ],
      rpm: 312, errRate: 2.1, p95: 3400, attached: 3
    },
    {
      id: "chat-eu", name: "chat-eu",
      strategy: "latency",
      desc: "EU-residency traffic. Latency-based steering.",
      members: [
        { model: "mistral-large", logo: "mistral", weight: 100, role: "active" },
      ],
      rpm: 2104, errRate: 0.2, p95: 690, attached: 2
    },
    {
      id: "embeddings", name: "embeddings",
      strategy: "single",
      desc: "Vector embeddings via Bedrock.",
      members: [
        { model: "bedrock/llama-3.3-70b", logo: "bedrock", weight: 100, role: "active" },
      ],
      rpm: 5402, errRate: 0.0, p95: 410, attached: 6
    },
  ];

  const keys = [
    { id: "key_4f9a", prefix: "sk-lg-prod-7Q2mF…",  name: "checkout-service",      team: "Payments",  budget: 4500, budgetCap: 5000, rpm: 600, status: "active",    created: "2026-01-14", lastUsed: "2m ago" },
    { id: "key_2c81", prefix: "sk-lg-prod-XnB91…",  name: "support-bot",           team: "Support",   budget: 312,  budgetCap: 1000, rpm: 200, status: "active",    created: "2026-02-04", lastUsed: "12s ago" },
    { id: "key_88aa", prefix: "sk-lg-stg-9Fp02…",   name: "ml-research",           team: "Research",  budget: 9870, budgetCap: 10000, rpm: 1200,status: "throttled", created: "2025-11-22", lastUsed: "live" },
    { id: "key_71d3", prefix: "sk-lg-prod-A7zKw…",  name: "internal-rag-eval",     team: "Platform",  budget: 88,   budgetCap: 500,  rpm: 100, status: "active",    created: "2026-03-09", lastUsed: "4h ago" },
    { id: "key_55fb", prefix: "sk-lg-prod-pQ4dH…",  name: "marketing-content",     team: "Growth",    budget: 1450, budgetCap: 1500, rpm: 60,  status: "warning",   created: "2026-01-30", lastUsed: "1m ago" },
    { id: "key_0a23", prefix: "sk-lg-stg-Z0w7R…",   name: "elliott-laptop",        team: "Platform",  budget: 12,   budgetCap: 100,  rpm: 30,  status: "active",    created: "2026-04-18", lastUsed: "3d ago" },
    { id: "key_e904", prefix: "sk-lg-prod-Mn82V…",  name: "legacy-batch",          team: "Data",      budget: 0,    budgetCap: 2000, rpm: 100, status: "disabled",  created: "2025-08-01", lastUsed: "94d ago" },
  ];

  const teams = [
    { id: "tm_pay",     name: "Payments",  members: 12, keys: 3, budget: 4500,  cap: 5000,  models: ["chat-prod"], owner: "tina@" },
    { id: "tm_supp",    name: "Support",   members: 6,  keys: 2, budget: 312,   cap: 1500,  models: ["chat-cheap"], owner: "rj@" },
    { id: "tm_research",name: "Research",  members: 22, keys: 5, budget: 9870,  cap: 12000, models: ["chat-long","chat-prod"], owner: "kavi@" },
    { id: "tm_growth",  name: "Growth",    members: 9,  keys: 4, budget: 1450,  cap: 2000,  models: ["chat-cheap"], owner: "mara@" },
    { id: "tm_data",    name: "Data",      members: 14, keys: 6, budget: 2104,  cap: 6000,  models: ["embeddings","chat-eu"], owner: "luca@" },
    { id: "tm_plat",    name: "Platform",  members: 8,  keys: 7, budget: 312,   cap: 5000,  models: ["chat-prod","chat-cheap","embeddings"], owner: "elliott@" },
  ];

  const requests = [
    { id: "req_01HZ8Xabc1", t: "12:04:09.412", method: "chat.completions", model: "claude-sonnet-4-5", route: "chat-prod", key: "checkout-service", team: "Payments", status: 200, latency: 1142, tokens: { in: 842, out: 312 }, cost: 0.00782, fb: false },
    { id: "req_01HZ8Xabd2", t: "12:04:08.901", method: "chat.completions", model: "gpt-4o-mini",        route: "chat-cheap",key: "support-bot",      team: "Support",  status: 200, latency: 612,  tokens: { in: 1240, out: 89 }, cost: 0.00072, fb: false },
    { id: "req_01HZ8Xabd3", t: "12:04:08.554", method: "chat.completions", model: "gpt-4o",             route: "chat-prod", key: "checkout-service", team: "Payments", status: 200, latency: 980,  tokens: { in: 514, out: 222 },  cost: 0.0035,  fb: true  },
    { id: "req_01HZ8Xabd4", t: "12:04:07.220", method: "embeddings",       model: "llama-3.3-70b",      route: "embeddings",key: "ml-research",      team: "Research", status: 200, latency: 410,  tokens: { in: 4202, out: 0 },   cost: 0.00302, fb: false },
    { id: "req_01HZ8Xabd5", t: "12:04:06.110", method: "chat.completions", model: "gemini-2.5-pro",     route: "chat-long", key: "ml-research",      team: "Research", status: 429, latency: 88,   tokens: { in: 0, out: 0 },      cost: 0,       fb: false },
    { id: "req_01HZ8Xabd6", t: "12:04:05.001", method: "chat.completions", model: "claude-haiku-4-5",   route: "chat-cheap",key: "marketing-content",team: "Growth",   status: 200, latency: 540,  tokens: { in: 612, out: 188 },  cost: 0.00038, fb: false },
    { id: "req_01HZ8Xabd7", t: "12:04:03.812", method: "chat.completions", model: "mistral-large-2",    route: "chat-eu",   key: "internal-rag-eval",team: "Platform", status: 200, latency: 690,  tokens: { in: 1820, out: 412 }, cost: 0.0061,  fb: false },
    { id: "req_01HZ8Xabd8", t: "12:04:02.440", method: "chat.completions", model: "claude-sonnet-4-5",  route: "chat-prod", key: "elliott-laptop",   team: "Platform", status: 500, latency: 4220, tokens: { in: 802, out: 0 },    cost: 0,       fb: true  },
    { id: "req_01HZ8Xabd9", t: "12:04:01.118", method: "chat.completions", model: "gpt-4o-mini",        route: "chat-cheap",key: "support-bot",      team: "Support",  status: 200, latency: 504,  tokens: { in: 220, out: 110 },  cost: 0.00012, fb: false },
    { id: "req_01HZ8Xabe0", t: "12:04:00.604", method: "chat.completions", model: "gemini-2.5-flash",   route: "chat-cheap",key: "marketing-content",team: "Growth",   status: 200, latency: 320,  tokens: { in: 412, out: 188 },  cost: 0.00009, fb: false },
    { id: "req_01HZ8Xabe1", t: "12:03:59.220", method: "chat.completions", model: "claude-sonnet-4-5",  route: "chat-prod", key: "checkout-service", team: "Payments", status: 200, latency: 1080, tokens: { in: 1220, out: 312 }, cost: 0.0083,  fb: false },
    { id: "req_01HZ8Xabe2", t: "12:03:58.001", method: "chat.completions", model: "gpt-4o",             route: "chat-prod", key: "internal-rag-eval",team: "Platform", status: 400, latency: 22,   tokens: { in: 0, out: 0 },      cost: 0,       fb: false },
  ];

  const guardrails = [
    { id: "gr_pii",     name: "PII redaction",          kind: "redact",  scope: "request",  hits24h: 942,  enabled: true,  owner: "Security" },
    { id: "gr_secrets", name: "Secret scanner",         kind: "block",   scope: "request",  hits24h: 12,   enabled: true,  owner: "Security" },
    { id: "gr_prompt",  name: "Prompt injection guard", kind: "block",   scope: "request",  hits24h: 318,  enabled: true,  owner: "Platform" },
    { id: "gr_tox",     name: "Toxic content filter",   kind: "warn",    scope: "response", hits24h: 47,   enabled: true,  owner: "Trust & Safety" },
    { id: "gr_jail",    name: "Jailbreak heuristics",   kind: "block",   scope: "request",  hits24h: 88,   enabled: false, owner: "Trust & Safety" },
    { id: "gr_topic",   name: "Off-topic dropper",      kind: "warn",    scope: "request",  hits24h: 0,    enabled: false, owner: "Support" },
  ];

  const audit = [
    { t: "12:02:14", who: "elliott@", action: "key.create",    target: "key_0a23 (elliott-laptop)", ip: "192.0.2.4" },
    { t: "11:58:01", who: "tina@",    action: "route.update",  target: "chat-prod (added fallback gpt-4o)", ip: "192.0.2.42" },
    { t: "11:42:33", who: "system",   action: "provider.degrade", target: "Vertex · gemini (p95 > 2.0s)", ip: "—" },
    { t: "10:21:47", who: "kavi@",    action: "budget.update", target: "Research cap 10k → 12k", ip: "192.0.2.18" },
    { t: "09:14:08", who: "elliott@", action: "guardrail.toggle", target: "Jailbreak heuristics → off (test)", ip: "192.0.2.4" },
    { t: "08:47:12", who: "rj@",      action: "key.rotate",    target: "key_2c81 (support-bot)", ip: "192.0.2.71" },
    { t: "08:02:00", who: "system",   action: "ratelimit.trip",target: "key_88aa (ml-research) 1200rpm", ip: "—" },
    { t: "yesterday",who: "luca@",    action: "team.create",   target: "Data", ip: "192.0.2.66" },
  ];

  // 24h sparkline data — points 0..n
  function spark(n, base, jitter, slope = 0) {
    const out = [];
    for (let i = 0; i < n; i++) out.push(Math.max(0, base + (Math.random() - 0.5) * jitter + slope * i));
    return out;
  }

  const dash = {
    requests24h: 1240821,
    requestsDelta: 8.4,
    spend24h: 1284.42,
    spendDelta: -2.1,
    errors24h: 0.42,
    errorsDelta: -18,
    p95: 1180,
    p95Delta: 4.2,
    requestsSpark: spark(48, 26000, 6000, 80),
    spendSpark:    spark(48, 26.4, 6, 0.05),
    errorsSpark:   spark(48, 0.4, 0.18),
    p95Spark:      spark(48, 1180, 240, -4),
  };

  return { providers, models, routes, keys, teams, requests, guardrails, audit, dash };
})();
