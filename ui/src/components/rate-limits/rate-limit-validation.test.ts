import { describe, it, expect } from 'vitest';
import { validateRateLimitRule } from './rate-limit-validation';

describe('rate limit rule validation', () => {
  it('rejects negative rps', () => {
    const errors = validateRateLimitRule({ scope: 'model', scope_id: 'gpt-4o', rps: -1 });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('rejects negative tpm', () => {
    const errors = validateRateLimitRule({ scope: 'model', scope_id: 'gpt-4o', rps: 10, tpm: -1 });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('rejects invalid scope', () => {
    const errors = validateRateLimitRule({ scope: 'invalid' as 'model', scope_id: 'x', rps: 10 });
    expect(errors.length).toBeGreaterThan(0);
  });

  it('accepts valid model-scope rule', () => {
    const errors = validateRateLimitRule({ scope: 'model', scope_id: 'gpt-4o', rps: 100, tpm: 60000 });
    expect(errors).toHaveLength(0);
  });
});
