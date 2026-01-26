-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_members
    ADD COLUMN left_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_members
    DROP COLUMN left_at;
-- +goose StatementEnd
