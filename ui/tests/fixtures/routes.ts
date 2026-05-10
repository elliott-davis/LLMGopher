import type { RoutePolicy } from '../../src/lib/admin-surface-contracts';

export const routes: RoutePolicy[] = [
  {
    id: 'route_gpt4o_fallback',
    model_alias: 'gpt-4o',
    strategy: 'fallback',
    enabled: true,
    targets: [
      { provider_id: 'prov_openai_prod', provider_name: 'OpenAI Primary', weight: 1, order: 0, health_state: 'ok' },
      { provider_id: 'prov_anthropic_degraded', provider_name: 'Anthropic Fallback', weight: 1, order: 1, health_state: 'degraded' },
    ],
    primary_provider_id: 'prov_openai_prod',
    fallback_provider_ids: ['prov_anthropic_degraded'],
  },
  {
    id: 'route_gpt4o_mini_weighted',
    model_alias: 'gpt-4o-mini',
    strategy: 'weighted',
    enabled: true,
    targets: [
      { provider_id: 'prov_openai_prod', provider_name: 'OpenAI Primary', weight: 70, health_state: 'ok' },
      { provider_id: 'prov_anthropic_degraded', provider_name: 'Anthropic Secondary', weight: 30, health_state: 'ok' },
    ],
  },
  {
    id: 'route_claude_latency',
    model_alias: 'claude-3.5-sonnet',
    strategy: 'latency',
    enabled: true,
    targets: [
      { provider_id: 'prov_openai_prod', provider_name: 'OpenAI', weight: 1, latency_ms: 250, health_state: 'ok' },
      { provider_id: 'prov_anthropic_degraded', provider_name: 'Anthropic', weight: 1, latency_ms: 180, health_state: 'ok' },
    ],
  },
  {
    id: 'route_haiku_single',
    model_alias: 'claude-3-haiku',
    strategy: 'single',
    enabled: true,
    targets: [
      { provider_id: 'prov_anthropic_degraded', provider_name: 'Anthropic', weight: 1, health_state: 'ok' },
    ],
  },
];
