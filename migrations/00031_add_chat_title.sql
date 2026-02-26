-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN title TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN title;
-- +goose StatementEnd
