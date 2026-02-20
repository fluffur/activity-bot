-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN command_prefix TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN command_prefix;
-- +goose StatementEnd
