-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_members
    RENAME COLUMN exempt_until TO rest_until;

ALTER TABLE exempt_requests
    RENAME COLUMN exempt_until TO rest_until;

ALTER TABLE exempt_requests
    RENAME TO rest_requests;

ALTER TYPE exempt_status RENAME TO rest_status;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat_members
    RENAME COLUMN rest_until TO exempt_until;

ALTER TABLE rest_requests
    RENAME COLUMN rest_until TO exempt_until;

ALTER TABLE rest_requests
    RENAME TO exempt_requests;

ALTER TYPE rest_status RENAME TO exempt_status;
-- +goose StatementEnd
