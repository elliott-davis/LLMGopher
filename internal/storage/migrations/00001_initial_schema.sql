-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS providers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL UNIQUE,
    base_url varchar(1024) NOT NULL,
    auth_type varchar(64) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS models (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id uuid NOT NULL REFERENCES providers (id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    alias varchar(255) NOT NULL UNIQUE,
    context_window integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, name)
);

CREATE INDEX IF NOT EXISTS idx_models_provider_id ON models (provider_id);

CREATE TABLE IF NOT EXISTS api_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash varchar(255) NOT NULL UNIQUE,
    name varchar(255) NOT NULL,
    rate_limit_rps integer NOT NULL DEFAULT 0,
    is_active boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    api_key_id TEXT NOT NULL,
    model TEXT NOT NULL,
    provider TEXT NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    status_code INTEGER NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    streaming BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_api_key_id ON audit_log (api_key_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at);

CREATE TABLE IF NOT EXISTS api_key_budgets (
    api_key_id TEXT PRIMARY KEY,
    budget_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    spent_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS model_pricing (
    model_prefix TEXT PRIMARY KEY,
    prompt_per_1k DOUBLE PRECISION NOT NULL,
    completion_per_1k DOUBLE PRECISION NOT NULL,
    source TEXT NOT NULL DEFAULT 'seed',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_pricing_updated_at ON model_pricing (updated_at);

INSERT INTO api_key_budgets (api_key_id, budget_usd, spent_usd)
VALUES ('key-001', 100.00, 0.00)
ON CONFLICT (api_key_id) DO NOTHING;

INSERT INTO model_pricing (model_prefix, prompt_per_1k, completion_per_1k, source) VALUES
    ('gpt-4o',                        0.0025,    0.0100,   'seed'),
    ('gpt-4o-mini',                   0.00015,   0.0006,   'seed'),
    ('gpt-4-turbo',                   0.0100,    0.0300,   'seed'),
    ('gpt-4',                         0.0300,    0.0600,   'seed'),
    ('gpt-3.5-turbo',                 0.0005,    0.0015,   'seed'),
    ('o1',                            0.0150,    0.0600,   'seed'),
    ('o1-mini',                       0.0030,    0.0120,   'seed'),
    ('o3-mini',                       0.0011,    0.0044,   'seed'),
    ('claude-3-5-sonnet',             0.0030,    0.0150,   'seed'),
    ('claude-3-5-haiku',              0.0008,    0.0040,   'seed'),
    ('claude-3-opus',                 0.0150,    0.0750,   'seed'),
    ('claude-3-sonnet',               0.0030,    0.0150,   'seed'),
    ('claude-3-haiku',                0.00025,   0.00125,  'seed'),
    ('gemini/gemini-2.5-pro',         0.00125,   0.0100,   'seed'),
    ('gemini/gemini-2.5-flash',       0.00015,   0.00060,  'seed'),
    ('gemini/gemini-2.0-flash',       0.00010,   0.00040,  'seed'),
    ('gemini/gemini-1.5-pro',         0.00125,   0.00500,  'seed'),
    ('gemini/gemini-1.5-flash',       0.000075,  0.00030,  'seed'),
    ('vertex_ai/gemini-2.5-pro',      0.00125,   0.0100,   'seed'),
    ('vertex_ai/gemini-2.5-flash',    0.00015,   0.00060,  'seed'),
    ('vertex_ai/gemini-2.0-flash',    0.00010,   0.00040,  'seed'),
    ('vertex_ai/gemini-1.5-pro',      0.00125,   0.00500,  'seed'),
    ('vertex_ai/gemini-1.5-flash',    0.000075,  0.00030,  'seed')
ON CONFLICT (model_prefix) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS model_pricing;
DROP TABLE IF EXISTS api_key_budgets;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS providers;
-- +goose StatementEnd
