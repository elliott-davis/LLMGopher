import { describe, it, expect } from 'vitest';
import { parseAuditFilter } from '@/lib/query-state';

describe('audit filters', () => {
  it('parseAuditFilter returns defaults for empty params', () => {
    const result = parseAuditFilter(new URLSearchParams());
    expect(result.actor).toBeUndefined();
  });
});
