-- +goose Up
-- +goose StatementBegin
ALTER TABLE api_key_budgets
    ADD COLUMN IF NOT EXISTS alert_threshold_pct INTEGER,
    ADD COLUMN IF NOT EXISTS budget_duration TEXT,
    ADD COLUMN IF NOT EXISTS budget_reset_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_alerted_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE api_key_budgets
    DROP COLUMN IF EXISTS last_alerted_at,
    DROP COLUMN IF EXISTS budget_reset_at,
    DROP COLUMN IF EXISTS budget_duration,
    DROP COLUMN IF EXISTS alert_threshold_pct;
-- +goose StatementEnd
