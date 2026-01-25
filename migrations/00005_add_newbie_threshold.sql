-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN newbie_threshold_days INT NOT NULL DEFAULT 3;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN newbie_threshold_days;
-- +goose StatementEnd
