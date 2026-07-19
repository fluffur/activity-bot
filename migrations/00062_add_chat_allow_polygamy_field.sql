-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN allow_polygamy boolean NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN allow_polygamy;
-- +goose StatementEnd