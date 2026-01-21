-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_members
    ADD COLUMN custom_title TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_members
    DROP COLUMN custom_title;
-- +goose StatementEnd
