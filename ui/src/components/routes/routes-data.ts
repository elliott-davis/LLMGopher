import type { RoutePolicy } from '@/lib/admin-surface-contracts';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function fetchRoutes(): Promise<{ routes: RoutePolicy[]; unavailable: boolean }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/routes`, { cache: 'no-store' });
    if (!res.ok) return { routes: [], unavailable: true };
    const body = await res.json() as { data: RoutePolicy[] };
    return { routes: body.data ?? [], unavailable: false };
  } catch {
    return { routes: [], unavailable: true };
  }
}
