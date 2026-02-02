-- +goose Up
-- +goose NO TRANSACTION
DROP INDEX CONCURRENTLY idx_messages_chat_date;

-- +goose Down
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY idx_messages_chat_date
    ON messages (chat_id, created_at);
