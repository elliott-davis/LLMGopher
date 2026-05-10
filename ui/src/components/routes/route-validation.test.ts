import { describe, it, expect } from 'vitest';
import { validateRoutePolicy } from './route-validation';

describe('route validation', () => {
  it('rejects weighted route with zero total weight', () => {
    const errors = validateRoutePolicy({ strategy: 'weighted', targets: [{ provider_id: 'p1', provider_name: 'P1', weight: 0 }] });
    expect(errors).toContain('Weighted route total weight must be greater than zero.');
  });

  it('rejects weighted route with negative weight', () => {
    const errors = validateRoutePolicy({ strategy: 'weighted', targets: [{ provider_id: 'p1', provider_name: 'P1', weight: -1 }] });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('rejects fallback route missing primary provider', () => {
    const errors = validateRoutePolicy({ strategy: 'fallback', targets: [], primary_provider_id: undefined });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('accepts valid single-provider route', () => {
    const errors = validateRoutePolicy({ strategy: 'single', targets: [{ provider_id: 'p1', provider_name: 'P1', weight: 1 }] });
    expect(errors).toHaveLength(0);
  });
});
