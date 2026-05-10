import { describe, expect, it } from 'vitest';

import {
  filterToSearchString,
  parseAuditFilter,
  parseLogsFilter,
  serializeAuditFilter,
  serializeLogsFilter,
} from './query-state';

describe('parseLogsFilter', () => {
  it('parses status from search params', () => {
    const params = new URLSearchParams({ status: '4xx' });
    const filter = parseLogsFilter(params);
    expect(filter.status).toBe('4xx');
  });

  it('parses status from record', () => {
    const filter = parseLogsFilter({ status: '2xx', page: '2', limit: '10' });
    expect(filter.status).toBe('2xx');
    expect(filter.page).toBe(2);
    expect(filter.limit).toBe(10);
  });

  it('returns defaults for empty params', () => {
    const filter = parseLogsFilter(new URLSearchParams());
    expect(filter.status).toBeUndefined();
    expect(filter.page).toBeUndefined();
    expect(filter.limit).toBeUndefined();
  });
});

describe('serializeLogsFilter', () => {
  it('round-trips status filter', () => {
    const filter = { status: '5xx' as const, page: 3, limit: 25 };
    const params = serializeLogsFilter(filter);
    const reparsed = parseLogsFilter(params);
    expect(reparsed.status).toBe('5xx');
    expect(reparsed.page).toBe(3);
    expect(reparsed.limit).toBe(25);
  });
});

describe('parseAuditFilter', () => {
  it('parses actor, action, from, to', () => {
    const params = new URLSearchParams({
      actor: 'key-001',
      action: 'chat.completion',
      from: '2026-05-01',
      to: '2026-05-09',
    });
    const filter = parseAuditFilter(params);
    expect(filter.actor).toBe('key-001');
    expect(filter.action).toBe('chat.completion');
    expect(filter.from).toBe('2026-05-01');
    expect(filter.to).toBe('2026-05-09');
  });

  it('returns empty filter for empty params', () => {
    const filter = parseAuditFilter(new URLSearchParams());
    expect(filter.actor).toBeUndefined();
    expect(filter.action).toBeUndefined();
    expect(filter.from).toBeUndefined();
    expect(filter.to).toBeUndefined();
  });
});

describe('serializeAuditFilter', () => {
  it('round-trips audit filters', () => {
    const filter = {
      actor: 'key-002',
      action: 'embeddings',
      from: '2026-04-01',
      to: '2026-04-30',
      limit: 50,
      offset: 100,
    };
    const params = serializeAuditFilter(filter);
    const reparsed = parseAuditFilter(params);
    expect(reparsed.actor).toBe('key-002');
    expect(reparsed.action).toBe('embeddings');
    expect(reparsed.from).toBe('2026-04-01');
    expect(reparsed.to).toBe('2026-04-30');
    expect(reparsed.limit).toBe(50);
    expect(reparsed.offset).toBe(100);
  });
});

describe('filterToSearchString', () => {
  it('returns empty string for empty params', () => {
    expect(filterToSearchString(new URLSearchParams())).toBe('');
  });

  it('returns ?-prefixed string for populated params', () => {
    const params = new URLSearchParams({ status: '2xx' });
    expect(filterToSearchString(params)).toBe('?status=2xx');
  });
});
