-- +goose Up
ALTER TABLE chats
    ADD COLUMN username_changed_notify_status SMALLINT NOT NULL DEFAULT 6;

-- +goose Down
ALTER TABLE chats
    DROP COLUMN username_changed_notify_status;