-- +goose Up
-- +goose StatementBegin
CREATE TABLE imported_activity
(
    chat_id        BIGINT      NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id        BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    period_start   TIMESTAMPTZ NOT NULL,
    period_end     TIMESTAMPTZ NOT NULL,
    messages_count INT         NOT NULL,

    PRIMARY KEY (chat_id, user_id, period_start, period_end),
    CHECK (period_end > period_start)
);

CREATE INDEX idx_imported_activity_period
    ON imported_activity (chat_id, period_start, period_end);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE imported_activity;
-- +goose StatementEnd
