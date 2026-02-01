-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_members
    RENAME COLUMN role TO status;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_members
    RENAME COLUMN status TO role;
-- +goose StatementEnd
