import { ProvidersView } from "@/components/providers/ProvidersView";
import type { Model, Provider } from "@/lib/types";

const GATEWAY = process.env.LLMGOPHER_GATEWAY_BASE ?? "http://gateway:8080";

async function fetchJSON<T>(url: string): Promise<T | null> {
  try {
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) return null;
    return res.json() as Promise<T>;
  } catch {
    return null;
  }
}

export default async function ProvidersPage() {
  const [providersResponse, modelsResponse] = await Promise.all([
    fetchJSON<Provider[]>(`${GATEWAY}/v1/admin/providers`),
    fetchJSON<Model[]>(`${GATEWAY}/v1/admin/models`),
  ]);

  const providers = Array.isArray(providersResponse) ? providersResponse : [];
  const modelCount = Array.isArray(modelsResponse) ? modelsResponse.length : 0;

  return (
    <ProvidersView
      providers={providers}
      modelCount={modelCount}
    />
  );
}
