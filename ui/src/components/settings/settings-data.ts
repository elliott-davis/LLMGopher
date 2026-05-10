import type { SettingCard } from '@/lib/admin-surface-contracts';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function fetchSettings(): Promise<{ cards: SettingCard[]; unavailable: boolean }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/settings`, { cache: 'no-store' });
    if (!res.ok) return { cards: [], unavailable: true };
    const body = await res.json() as { data: SettingCard[] };
    return { cards: body.data ?? [], unavailable: false };
  } catch {
    return { cards: [], unavailable: true };
  }
}

export async function saveDisplaySettings(values: Record<string, string>): Promise<{ ok: boolean; error?: string }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/settings/display`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
      cache: 'no-store',
    });
    if (!res.ok) return { ok: false, error: 'Save failed. Please try again.' };
    return { ok: true };
  } catch {
    return { ok: false, error: 'Service unavailable.' };
  }
}
