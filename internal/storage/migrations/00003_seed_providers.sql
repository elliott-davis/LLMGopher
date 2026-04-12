-- +goose Up
-- +goose StatementBegin
INSERT INTO providers (name, base_url, auth_type) VALUES
    ('OpenAI',         'https://api.openai.com',    'bearer'),
    ('Anthropic',      'https://api.anthropic.com', 'x-api-key'),
    ('Google Vertex',  '',                           'google_adc')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM providers WHERE name IN ('OpenAI', 'Anthropic', 'Google Vertex');
-- +goose StatementEnd
