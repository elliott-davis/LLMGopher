import type { AuditRecord } from '@/lib/types';
import type { AuditFilter, AuditListResponse } from '@/lib/admin-surface-contracts';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function fetchAuditEntries(filter: AuditFilter): Promise<{ entries: AuditRecord[]; total: number; unavailable: boolean }> {
  try {
    const params = new URLSearchParams();
    if (filter.actor) params.set('actor', filter.actor);
    if (filter.action) params.set('action', filter.action);
    if (filter.from) params.set('from', filter.from);
    if (filter.to) params.set('to', filter.to);
    if (filter.limit) params.set('limit', String(filter.limit));
    if (filter.offset) params.set('offset', String(filter.offset));
    const url = `${GATEWAY_BASE}/v1/admin/audit${params.size > 0 ? '?' + params : ''}`;
    const res = await fetch(url, { cache: 'no-store' });
    if (!res.ok) return { entries: [], total: 0, unavailable: true };
    const body = await res.json() as AuditListResponse;
    return { entries: body.data ?? [], total: body.total ?? 0, unavailable: false };
  } catch {
    return { entries: [], total: 0, unavailable: true };
  }
}
