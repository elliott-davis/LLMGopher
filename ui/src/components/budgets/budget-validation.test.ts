import { describe, it, expect } from 'vitest';
import { validateBudgetPolicy } from './budget-validation';

describe('budget policy validation', () => {
  it('rejects negative limit', () => {
    const errors = validateBudgetPolicy({ limit_usd: -1, usage_usd: 0, alert_threshold: 0.85 });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('rejects negative usage', () => {
    const errors = validateBudgetPolicy({ limit_usd: 100, usage_usd: -1, alert_threshold: 0.85 });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('rejects alert threshold > 1', () => {
    const errors = validateBudgetPolicy({ limit_usd: 100, usage_usd: 50, alert_threshold: 1.5 });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('rejects alert threshold <= 0', () => {
    const errors = validateBudgetPolicy({ limit_usd: 100, usage_usd: 50, alert_threshold: 0 });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('accepts valid budget', () => {
    const errors = validateBudgetPolicy({ limit_usd: 100, usage_usd: 50, alert_threshold: 0.85 });
    expect(errors).toHaveLength(0);
  });
});
