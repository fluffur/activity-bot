-- +goose Up
ALTER TABLE chats
    ADD COLUMN emojis_enabled boolean NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE chats DROP COLUMN emojis_enabled;
