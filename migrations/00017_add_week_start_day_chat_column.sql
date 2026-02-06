-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD week_start_day SMALLINT DEFAULT 1 NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP week_start_day;
-- +goose StatementEnd
