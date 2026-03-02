-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_members
    ADD COLUMN rest_reason TEXT;
ALTER TABLE rest_requests ADD COLUMN reason TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_members
    DROP COLUMN rest_reason;
ALTER TABLE rest_requests DROP reason;
-- +goose StatementEnd