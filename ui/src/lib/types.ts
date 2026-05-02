export interface APIKey {
  id: string;
  key_hash: string;
  name: string;
  rate_limit_rps: number;
  is_active: boolean;
  expires_at?: string | null;
  metadata?: Record<string, string> | null;
  allowed_models?: string[] | null;
  created_at: string;
  updated_at: string;
}

export interface Model {
  id: string;
  provider_id: string;
  name: string;
  alias: string;
  context_window: number;
  created_at: string;
  updated_at: string;
}

export interface APIKeyFormValues {
  name: string;
  rate_limit_rps: number;
  expires_at: string | null;
  metadata: Record<string, string>;
  allowed_models: string[];
  is_active?: boolean;
}

export type APIKeyMutationResult =
  | { success: true; api_key?: string }
  | { success: false; error: string };

export interface GatewayErrorEnvelope {
  error?: {
    message?: string;
    type?: string;
    code?: string;
  } | string;
  message?: string;
}

export interface Provider {
  id: string;
  name: string;
  base_url: string;
  auth_type: string;
  has_credentials?: boolean;
  created_at?: string;
  updated_at?: string;
}
