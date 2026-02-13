-- +goose Up
ALTER TABLE messages
    ADD COLUMN id BIGSERIAL;

ALTER TABLE messages
    ALTER COLUMN id SET NOT NULL;

ALTER TABLE messages
    DROP CONSTRAINT messages_pkey;

ALTER TABLE messages
    ADD PRIMARY KEY (id);

ALTER TABLE messages
    ADD COLUMN message_id BIGINT;

CREATE UNIQUE INDEX uniq_messages_chat_message
    ON messages (chat_id, message_id)
    WHERE message_id IS NOT NULL AND message_id <> 0;

-- +goose Down
DROP INDEX IF EXISTS uniq_messages_chat_message;

ALTER TABLE messages
    DROP CONSTRAINT messages_pkey;

ALTER TABLE messages
    DROP COLUMN id;

ALTER TABLE messages
    DROP COLUMN message_id;

ALTER TABLE messages
    ADD PRIMARY KEY (chat_id, user_id, created_at);
