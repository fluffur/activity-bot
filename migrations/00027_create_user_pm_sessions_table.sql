-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_pm_sessions
(
    user_id        BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    target_chat_id BIGINT NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_pm_sessions;
-- +goose StatementEnd
