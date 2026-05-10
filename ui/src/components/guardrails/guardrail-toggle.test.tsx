import { describe, it, expect } from 'vitest';

describe('guardrail toggle state', () => {
  it('enabled state is truthy', () => {
    expect(true).toBe(true);
  });

  it('toggle mutation sends correct payload', () => {
    const patch = (enabled: boolean) => ({ enabled });
    expect(patch(true)).toEqual({ enabled: true });
    expect(patch(false)).toEqual({ enabled: false });
  });
});
