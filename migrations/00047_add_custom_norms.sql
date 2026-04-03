-- +goose Up
ALTER TABLE chats ADD COLUMN bot_removed_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE chats DROP COLUMN bot_removed_at;
