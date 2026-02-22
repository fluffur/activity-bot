-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD COLUMN mentions_per_message INT NOT NULL DEFAULT 5,
    ADD COLUMN mention_types        INT NOT NULL DEFAULT 2; -- Default: Name (bit 1)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN mentions_per_message,
    DROP COLUMN mention_types;
-- +goose StatementEnd
