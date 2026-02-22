-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN allow_prefixless BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN allow_prefixless;
-- +goose StatementEnd
