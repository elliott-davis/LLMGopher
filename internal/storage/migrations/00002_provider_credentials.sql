-- +goose Up
-- +goose StatementBegin
ALTER TABLE providers
    ADD COLUMN IF NOT EXISTS credential_file_name TEXT,
    ADD COLUMN IF NOT EXISTS credential_ciphertext BYTEA,
    ADD COLUMN IF NOT EXISTS credential_nonce BYTEA,
    ADD COLUMN IF NOT EXISTS has_credentials BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE providers
    DROP COLUMN IF EXISTS has_credentials,
    DROP COLUMN IF EXISTS credential_nonce,
    DROP COLUMN IF EXISTS credential_ciphertext,
    DROP COLUMN IF EXISTS credential_file_name;
-- +goose StatementEnd
