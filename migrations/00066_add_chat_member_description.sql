-- +goose Up
ALTER TABLE chat_members ADD COLUMN description TEXT;

-- +goose Down
ALTER TABLE chat_members DROP COLUMN description;
