import { NextRequest } from 'next/server';

const GATEWAY_BASE = process.env.LLMGOPHER_GATEWAY_BASE ?? 'http://gateway:8080';

export async function GET(_req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/admin/logs/${id}`, { cache: 'no-store' });
    const body = await res.json();
    return Response.json(body, { status: res.status });
  } catch {
    return Response.json({ error: { message: 'unavailable', type: 'service_error' } }, { status: 503 });
  }
}
