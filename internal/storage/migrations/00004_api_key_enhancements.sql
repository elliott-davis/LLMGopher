-- +goose Up
-- +goose StatementBegin
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS allowed_models TEXT[];
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE api_keys
    DROP COLUMN IF EXISTS allowed_models,
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS expires_at;
-- +goose StatementEnd
