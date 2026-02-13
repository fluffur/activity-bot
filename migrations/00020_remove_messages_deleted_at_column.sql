-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages
    DROP COLUMN deleted_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages
    ADD deleted_at TIMESTAMPTZ;
-- +goose StatementEnd
