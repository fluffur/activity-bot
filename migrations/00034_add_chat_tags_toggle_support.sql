-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD COLUMN tags_enabled BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN tags_enabled;
-- +goose StatementEnd
