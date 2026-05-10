import type { GuardrailRule } from '@/lib/admin-surface-contracts';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function fetchGuardrails(): Promise<{ guardrails: GuardrailRule[]; unavailable: boolean }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/guardrails`, { cache: 'no-store' });
    if (!res.ok) return { guardrails: [], unavailable: true };
    const body = await res.json() as { data: GuardrailRule[] };
    return { guardrails: body.data ?? [], unavailable: false };
  } catch {
    return { guardrails: [], unavailable: true };
  }
}

export async function toggleGuardrail(id: string, enabled: boolean): Promise<{ ok: boolean; error?: string }> {
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/guardrails/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
      cache: 'no-store',
    });
    if (!res.ok) return { ok: false, error: 'Toggle failed. Please try again.' };
    return { ok: true };
  } catch {
    return { ok: false, error: 'Service unavailable.' };
  }
}
