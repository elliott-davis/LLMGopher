import type { RoutePolicy } from '@/lib/admin-surface-contracts';

export function validateRoutePolicy(route: {
  strategy: string;
  targets: Array<{ provider_id: string; provider_name: string; weight: number }>;
  primary_provider_id?: string;
}): string[] {
  const errors: string[] = [];
  if (route.strategy === 'weighted') {
    const hasNegative = route.targets.some((t) => t.weight < 0);
    if (hasNegative) errors.push('Weights must be non-negative.');
    const total = route.targets.reduce((sum, t) => sum + t.weight, 0);
    if (total <= 0) errors.push('Weighted route total weight must be greater than zero.');
  }
  if (route.strategy === 'fallback') {
    if (!route.primary_provider_id && route.targets.length === 0) {
      errors.push('Fallback route requires a primary provider.');
    }
    if (route.targets.length < 2 && !route.primary_provider_id) {
      errors.push('Fallback route requires a primary provider.');
    }
  }
  return errors;
}
