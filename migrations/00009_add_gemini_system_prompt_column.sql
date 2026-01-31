-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD COLUMN gemini_system_prompt TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN gemini_system_prompt;
-- +goose StatementEnd
