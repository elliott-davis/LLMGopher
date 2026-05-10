import { describe, it, expect } from 'vitest';
import { parseLogsFilter } from '@/lib/query-state';

describe('logs filters', () => {
  it('parseLogsFilter returns defaults for empty params', () => {
    const result = parseLogsFilter(new URLSearchParams());
    expect(result.status).toBeUndefined();
  });
});
