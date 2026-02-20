-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN moderation_enabled BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN moderation_enabled;
-- +goose StatementEnd
