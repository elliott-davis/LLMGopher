import { describe, expect, it } from 'vitest';

import {
  formatCurrency,
  formatLatency,
  formatPercentage,
  formatStatusLabel,
  formatTokenCount,
  unavailableCopy,
  utilizationState,
} from './admin-surface-format';

describe('formatStatusLabel', () => {
  it('formats 200', () => expect(formatStatusLabel(200)).toBe('200 OK'));
  it('formats 404', () => expect(formatStatusLabel(404)).toBe('404 Not Found'));
  it('formats 500', () => expect(formatStatusLabel(500)).toBe('500 Internal Server Error'));
  it('formats 429', () => expect(formatStatusLabel(429)).toBe('429 Too Many Requests'));
});

describe('formatLatency', () => {
  it('returns <1ms for 0', () => expect(formatLatency(0)).toBe('<1ms'));
  it('returns Xms for sub-second values', () => expect(formatLatency(500)).toBe('500ms'));
  it('returns X.Xs for >= 1000ms', () => expect(formatLatency(1500)).toBe('1.5s'));
});

describe('formatCurrency', () => {
  it('formats $0', () => expect(formatCurrency(0)).toBe('$0.0000'));
  it('formats $0.0024', () => expect(formatCurrency(0.0024)).toBe('$0.0024'));
  it('formats $10.50', () => expect(formatCurrency(10.5)).toBe('$10.5000'));
});

describe('formatTokenCount', () => {
  it('formats under 1000', () => expect(formatTokenCount(500)).toBe('500'));
  it('formats thousands', () => expect(formatTokenCount(1500)).toBe('1.5k'));
  it('formats millions', () => expect(formatTokenCount(1_500_000)).toBe('1.5M'));
});

describe('formatPercentage', () => {
  it('formats 0', () => expect(formatPercentage(0)).toBe('0%'));
  it('formats 0.5', () => expect(formatPercentage(0.5)).toBe('50%'));
  it('formats 0.86', () => expect(formatPercentage(0.86)).toBe('86%'));
  it('formats 1.0', () => expect(formatPercentage(1.0)).toBe('100%'));
});

describe('utilizationState', () => {
  it('returns ok when below threshold', () => {
    expect(utilizationState(0.5, 0.8)).toBe('ok');
  });
  it('returns near_cap when at threshold', () => {
    expect(utilizationState(0.8, 0.8)).toBe('near_cap');
  });
  it('returns over_cap when ratio >= 1.0', () => {
    expect(utilizationState(1.0, 0.8)).toBe('over_cap');
    expect(utilizationState(1.2, 0.8)).toBe('over_cap');
  });
});

describe('unavailableCopy', () => {
  it('returns non-empty string for any surface name', () => {
    const copy = unavailableCopy('Guardrails');
    expect(copy.length).toBeGreaterThan(0);
    expect(copy).toContain('Guardrails');
  });
});
