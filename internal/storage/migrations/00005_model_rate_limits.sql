-- +goose Up
-- +goose StatementBegin
ALTER TABLE models
    ADD COLUMN IF NOT EXISTS rate_limit_rps INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE models
    DROP COLUMN IF EXISTS rate_limit_rps;
-- +goose StatementEnd
