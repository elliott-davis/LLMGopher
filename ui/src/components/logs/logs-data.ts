import type { LogRow, LogDetail, LogsFilter, LogsListResponse } from '@/lib/admin-surface-contracts';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function fetchLogs(filter: LogsFilter): Promise<{ rows: LogRow[]; total: number; unavailable: boolean }> {
  try {
    const params = new URLSearchParams();
    if (filter.status && filter.status !== 'all') params.set('status', filter.status);
    if (filter.limit) params.set('limit', String(filter.limit));
    const url = `${GATEWAY_BASE}/v1/admin/logs${params.size > 0 ? '?' + params : ''}`;
    const res = await fetch(url, { cache: 'no-store' });
    if (!res.ok) return { rows: [], total: 0, unavailable: true };
    const body = await res.json() as LogsListResponse;
    const rows = body.data ?? [];
    const total = body.pagination?.total ?? rows.length;
    return { rows, total, unavailable: false };
  } catch {
    return { rows: [], total: 0, unavailable: true };
  }
}

export async function fetchLogDetail(id: string): Promise<{ detail: LogDetail | null; unavailable: boolean }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/logs/${id}`, { cache: 'no-store' });
    if (!res.ok) return { detail: null, unavailable: true };
    const detail = await res.json() as LogDetail;
    return { detail, unavailable: false };
  } catch {
    return { detail: null, unavailable: true };
  }
}

export function isStatusMatch(statusCode: number, filter: string): boolean {
  if (filter === '2xx') return statusCode >= 200 && statusCode < 300;
  if (filter === '4xx') return statusCode >= 400 && statusCode < 500;
  if (filter === '5xx') return statusCode >= 500 && statusCode < 600;
  return true;
}
