-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN emoji           TEXT,
    ADD COLUMN custom_emoji_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN emoji,
    DROP COLUMN custom_emoji_id;
-- +goose StatementEnd