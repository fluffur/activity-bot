-- +goose Up
-- +goose StatementBegin
ALTER TABLE bot_developers ADD COLUMN chat_id BIGINT;
DELETE FROM bot_developers; -- Clear existing global developers
ALTER TABLE bot_developers ALTER COLUMN chat_id SET NOT NULL;
ALTER TABLE bot_developers DROP CONSTRAINT bot_developers_pkey;
ALTER TABLE bot_developers ADD PRIMARY KEY (user_id, chat_id);
ALTER TABLE bot_developers ADD CONSTRAINT bot_developers_chat_id_fkey FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE bot_developers DROP CONSTRAINT bot_developers_chat_id_fkey;
ALTER TABLE bot_developers DROP CONSTRAINT bot_developers_pkey;
ALTER TABLE bot_developers DROP COLUMN chat_id;
ALTER TABLE bot_developers ADD PRIMARY KEY (user_id);
-- +goose StatementEnd
