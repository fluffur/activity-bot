-- +goose Up
ALTER TABLE chat_members ADD COLUMN emoji TEXT;

-- +goose Down
ALTER TABLE chat_members DROP COLUMN emoji;
