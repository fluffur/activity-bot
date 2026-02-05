-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats RENAME COLUMN gemini_system_prompt TO ai_system_prompt;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats RENAME COLUMN ai_system_prompt TO gemini_system_prompt    ;
-- +goose StatementEnd
