import { AuditFilter, LogsFilter } from './admin-surface-contracts';

type ParamSource = URLSearchParams | Record<string, string>;

function readParam(source: ParamSource, key: string): string {
  if (source instanceof URLSearchParams) {
    return source.get(key)?.trim() ?? '';
  }
  return String(source[key] ?? '').trim();
}

export function parseLogsFilter(searchParams: ParamSource): LogsFilter {
  const status = readParam(searchParams, 'status');
  const page = readParam(searchParams, 'page');
  const limit = readParam(searchParams, 'limit');

  const filter: LogsFilter = {};

  if (status === 'all' || status === '2xx' || status === '4xx' || status === '5xx' || status === 'fallback') {
    filter.status = status;
  }
  if (page) {
    const parsed = parseInt(page, 10);
    if (!isNaN(parsed) && parsed > 0) filter.page = parsed;
  }
  if (limit) {
    const parsed = parseInt(limit, 10);
    if (!isNaN(parsed) && parsed > 0) filter.limit = parsed;
  }

  return filter;
}

export function serializeLogsFilter(filter: LogsFilter): URLSearchParams {
  const params = new URLSearchParams();
  if (filter.status) params.set('status', filter.status);
  if (filter.page !== undefined) params.set('page', String(filter.page));
  if (filter.limit !== undefined) params.set('limit', String(filter.limit));
  return params;
}

export function parseAuditFilter(searchParams: ParamSource): AuditFilter {
  const filter: AuditFilter = {};

  const actor = readParam(searchParams, 'actor');
  const action = readParam(searchParams, 'action');
  const from = readParam(searchParams, 'from');
  const to = readParam(searchParams, 'to');
  const limit = readParam(searchParams, 'limit');
  const offset = readParam(searchParams, 'offset');

  if (actor) filter.actor = actor;
  if (action) filter.action = action;
  if (from) filter.from = from;
  if (to) filter.to = to;
  if (limit) {
    const parsed = parseInt(limit, 10);
    if (!isNaN(parsed) && parsed > 0) filter.limit = parsed;
  }
  if (offset) {
    const parsed = parseInt(offset, 10);
    if (!isNaN(parsed) && parsed >= 0) filter.offset = parsed;
  }

  return filter;
}

export function serializeAuditFilter(filter: AuditFilter): URLSearchParams {
  const params = new URLSearchParams();
  if (filter.actor) params.set('actor', filter.actor);
  if (filter.action) params.set('action', filter.action);
  if (filter.from) params.set('from', filter.from);
  if (filter.to) params.set('to', filter.to);
  if (filter.limit !== undefined) params.set('limit', String(filter.limit));
  if (filter.offset !== undefined) params.set('offset', String(filter.offset));
  return params;
}

export function filterToSearchString(params: URLSearchParams): string {
  const str = params.toString();
  return str ? `?${str}` : '';
}
