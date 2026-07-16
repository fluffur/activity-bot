-- +goose Up

DROP TABLE chat_rp_commands;

CREATE TABLE chat_rp_commands
(
    id         BIGSERIAL PRIMARY KEY,
    chat_id    BIGINT      NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    trigger    TEXT        NOT NULL,
    action     TEXT        NOT NULL,
    emojis     TEXT        NOT NULL DEFAULT '',
    created_by BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (chat_id, trigger)
);

-- +goose Down

DROP TABLE chat_rp_commands;

CREATE TABLE chat_rp_commands
(
    chat_id            BIGINT                                       NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    trigger            TEXT                                         NOT NULL,
    trigger_normalized TEXT                                         NOT NULL,
    template           TEXT,
    emoji_json         JSONB                    DEFAULT '[]'::jsonb NOT NULL,
    created_by         BIGINT                                       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at         TIMESTAMPTZ             DEFAULT now()         NOT NULL,
    updated_at         TIMESTAMPTZ             DEFAULT now()         NOT NULL,

    PRIMARY KEY (chat_id, trigger_normalized)
);