-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD COLUMN call_on_join         BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN welcome_call_message TEXT    NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN call_on_join,
    DROP COLUMN welcome_call_message;
-- +goose StatementEnd
