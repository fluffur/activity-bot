-- +goose Up
-- +goose StatementBegin
CREATE TYPE moderation_type AS ENUM ('kick', 'mute', 'ban', 'warn');

CREATE TABLE moderation_actions
(
    id         SERIAL PRIMARY KEY,
    type       moderation_type NOT NULL,
    chat_id    BIGINT          NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id    BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    mod_id     BIGINT          NOT NULL,
    reason     TEXT,
    created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    until_date TIMESTAMPTZ
);

CREATE INDEX idx_moderation_recent ON moderation_actions (chat_id, user_id, type, created_at);

ALTER TABLE chats
    ADD COLUMN max_warns INT NOT NULL DEFAULT 3;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN max_warns;
DROP TABLE moderation_actions;
DROP TYPE moderation_type;

CREATE TABLE bans
(
    id         SERIAL PRIMARY KEY,
    chat_id    BIGINT    NOT NULL,
    user_id    BIGINT    NOT NULL,
    reason     TEXT,
    banned_by  BIGINT    NOT NULL,
    banned_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    until_date TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);
CREATE INDEX idx_bans_chat_user ON bans (chat_id, user_id);
-- +goose StatementEnd
