export interface APIKey {
  id: string;
  key_hash: string;
  name: string;
  rate_limit_rps: number;
  is_active: boolean;
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

export interface Provider {
  id: string;
  name: string;
  base_url: string;
  auth_type: string;
  has_credentials?: boolean;
  created_at?: string;
  updated_at?: string;
}
