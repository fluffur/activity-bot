-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN broadcast_enabled BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN broadcast_enabled;
-- +goose StatementEnd