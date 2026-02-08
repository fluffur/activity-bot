-- +goose Up
-- +goose StatementBegin
CREATE TABLE bot_developers
(
    user_id BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    role    TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE bot_developers;
-- +goose StatementEnd
