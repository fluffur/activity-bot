-- +goose Up
ALTER TABLE chat_members ADD COLUMN exclude_from_call BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE chat_members DROP COLUMN exclude_from_call;
