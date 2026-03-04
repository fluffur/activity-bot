-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN week_start_time TIME NOT NULL DEFAULT '00:00';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN week_start_time;
-- +goose StatementEnd