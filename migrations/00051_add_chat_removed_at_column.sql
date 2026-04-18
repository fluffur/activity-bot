-- +goose Up
ALTER TABLE chats ADD COLUMN removed_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE chat_members DROP COLUMN removed_at;
