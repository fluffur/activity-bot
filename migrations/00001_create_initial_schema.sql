-- +goose Up
-- +goose StatementBegin
CREATE
    EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE chats
(
    id          BIGINT PRIMARY KEY,
    weekly_norm INT NOT NULL CHECK (weekly_norm >= 0)
);

CREATE TABLE users
(
    id         BIGINT PRIMARY KEY,
    username   TEXT,
    first_name TEXT,
    last_name  TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);


CREATE TABLE chat_members
(
    chat_id      BIGINT      NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id      BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    exempt_until TIMESTAMPTZ,
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messages
(
    chat_id    BIGINT      NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (chat_id, user_id, created_at)
);

CREATE TYPE exempt_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TABLE exempt_requests
(
    chat_id      BIGINT        NOT NULL,
    user_id      BIGINT        NOT NULL,
    requested_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    exempt_until TIMESTAMPTZ   NOT NULL,
    status       exempt_status NOT NULL DEFAULT 'pending',
    message_id   BIGINT, -- telegram message id
    PRIMARY KEY (chat_id, user_id, message_id),
    FOREIGN KEY (chat_id, user_id)
        REFERENCES chat_members (chat_id, user_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_exempt_requests_chat_user
    ON exempt_requests (chat_id, user_id);

CREATE INDEX idx_exempt_requests_pending
    ON exempt_requests (chat_id, user_id)
    WHERE status = 'pending';

CREATE INDEX idx_messages_chat_user_date
    ON messages (chat_id, user_id, created_at);

CREATE INDEX idx_messages_chat_date
    ON messages (chat_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE exempt_requests;
DROP TABLE messages;
DROP TABLE chat_members;
DROP TABLE users;
DROP TABLE chats;
-- +goose StatementEnd
