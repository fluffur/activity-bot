-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_members ADD COLUMN role TEXT NOT NULL DEFAULT 'administrator';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_members DROP COLUMN role;
-- +goose StatementEnd
