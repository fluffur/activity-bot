-- +goose Up
ALTER TABLE chat_members ADD COLUMN IF NOT EXISTS emoji_json JSONB;
ALTER TABLE users ADD COLUMN IF NOT EXISTS emoji_json JSONB;

-- +goose Down
ALTER TABLE chat_members DROP COLUMN IF EXISTS emoji_json;
ALTER TABLE users DROP COLUMN IF EXISTS emoji_json;