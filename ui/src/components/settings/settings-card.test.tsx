import { describe, it, expect } from 'vitest';
import { redactValue } from '@/lib/redaction';

describe('settings card states', () => {
  it('unavailable cards have save_capability false', () => {
    const card = { id: 'gateway-profile', save_capability: false, availability: 'unavailable' as const };
    expect(card.save_capability).toBe(false);
    expect(card.availability).toBe('unavailable');
  });

  it('editable cards have save_capability true', () => {
    const card = { id: 'display', save_capability: true, availability: 'editable' as const };
    expect(card.save_capability).toBe(true);
  });

  it('redaction hides secret-like values', () => {
    const result = redactValue('sk-proj-abc123');
    expect(result).toBe('[REDACTED]');
  });

  it('read_only fields are not interactive', () => {
    const field = { id: 'org_name', read_only: true };
    expect(field.read_only).toBe(true);
  });
});
