-- +goose Up
CREATE TABLE rewards
(
    id         BIGSERIAL PRIMARY KEY,
    chat_id    BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    author_id  BIGINT      NOT NULL,
    rank       SMALLINT    NOT NULL DEFAULT 0,
    reason     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (chat_id, user_id) REFERENCES chat_members (chat_id, user_id),
    FOREIGN KEY (chat_id, author_id) REFERENCES chat_members (chat_id, user_id)
);

CREATE INDEX rewards_chat_user_idx
    ON rewards (chat_id, user_id);

-- +goose Down
DROP TABLE rewards;