import type { RateLimitRule } from '@/lib/admin-surface-contracts';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function fetchRateLimits(): Promise<{ rules: RateLimitRule[]; unavailable: boolean }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/rate-limits`, { cache: 'no-store' });
    if (!res.ok) return { rules: [], unavailable: true };
    const body = await res.json() as { data: RateLimitRule[] };
    return { rules: body.data ?? [], unavailable: false };
  } catch {
    return { rules: [], unavailable: true };
  }
}
