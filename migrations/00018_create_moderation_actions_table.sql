-- +goose Up
-- +goose StatementBegin
CREATE TYPE moderation_type AS ENUM ('kick', 'mute', 'ban', 'warn');

CREATE TABLE moderation_actions
(
    id           BIGSERIAL PRIMARY KEY,
    type         moderation_type NOT NULL,
    chat_id      BIGINT          NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id      BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    moderator_id BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reason       TEXT,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    UNIQUE (chat_id, user_id, type, created_at)
);

CREATE INDEX idx_moderation_active
    ON moderation_actions (chat_id, user_id, type)
    WHERE revoked_at IS NULL;

ALTER TABLE chats
    ADD COLUMN max_warns INT NOT NULL DEFAULT 3;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN max_warns;

DROP TABLE moderation_actions;

DROP TYPE moderation_type;
-- +goose StatementEnd
