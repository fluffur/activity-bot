-- +goose Up
CREATE TABLE banned_users
(
    user_id    BIGINT PRIMARY KEY REFERENCES users(id),
    reason     TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);

CREATE TABLE banned_chats
(
    chat_id    BIGINT PRIMARY KEY REFERENCES chats(id),
    reason     TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);
-- +goose Down
DROP TABLE banned_chats;
DROP TABLE banned_users;