-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ALTER COLUMN mention_types SET DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    ALTER COLUMN mention_types SET DEFAULT 5;
-- +goose StatementEnd
