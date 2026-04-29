-- +goose Up
CREATE TABLE IF NOT EXISTS chat_rp_commands (
    chat_id BIGINT NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    trigger TEXT NOT NULL,
    trigger_normalized TEXT NOT NULL,
    template TEXT,
    emoji_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (chat_id, trigger_normalized)
);

CREATE INDEX IF NOT EXISTS idx_chat_rp_commands_chat_id
    ON chat_rp_commands (chat_id);

-- +goose Down
DROP TABLE IF EXISTS chat_rp_commands;
