-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD COLUMN skip_call_confirmation BOOL NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN skip_call_confirmation;
-- +goose StatementEnd
