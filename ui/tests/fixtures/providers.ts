import type { Provider } from "../../src/lib/types";

// Frozen "now" used across all fixtures: 2026-05-09T12:00:00.000Z
const NOW = "2026-05-09T12:00:00.000Z";

// Matches the real Provider interface from src/lib/types.ts.
// Extra fields (kind, health) are carried along for E2E tests that inspect them.
export interface ProviderFixture extends Provider {
  kind: "openai" | "anthropic" | "vertex" | "bedrock" | "cohere" | "generic";
  health: "healthy" | "degraded" | "offline";
}

export const providers: ProviderFixture[] = [
  {
    id: "prov_openai_prod",
    kind: "openai",
    name: "OpenAI · prod",
    base_url: "https://api.openai.com",
    auth_type: "api_key",
    has_credentials: true,
    health: "healthy",
    created_at: NOW,
    updated_at: NOW,
  },
  {
    id: "prov_anthropic_degraded",
    kind: "anthropic",
    name: "Anthropic · prod",
    base_url: "https://api.anthropic.com",
    auth_type: "api_key",
    has_credentials: true,
    health: "degraded",
    created_at: NOW,
    updated_at: NOW,
  },
  {
    id: "prov_vertex_offline",
    kind: "vertex",
    name: "Vertex AI · prod",
    base_url: "https://us-central1-aiplatform.googleapis.com",
    auth_type: "service_account",
    has_credentials: true,
    health: "offline",
    created_at: NOW,
    updated_at: NOW,
  },
];
