-- +goose Up
-- +goose StatementBegin
CREATE TABLE chat_admins
(
    chat_id    BIGINT      NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (chat_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chat_admins;
-- +goose StatementEnd
