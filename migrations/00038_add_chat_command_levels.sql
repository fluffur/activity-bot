-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat_members ADD COLUMN IF NOT EXISTS level SMALLINT NOT NULL DEFAULT 0;

-- Инициализируем уровни на основе текущих статусов
UPDATE chat_members SET level = 5 WHERE status = 'creator';
UPDATE chat_members SET level = 3 WHERE status = 'administrator';
UPDATE chat_members SET level = 0 WHERE status NOT IN ('creator', 'administrator');

CREATE TABLE IF NOT EXISTS chat_command_levels (
    chat_id BIGINT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    command_id TEXT NOT NULL,
    level SMALLINT NOT NULL,
    PRIMARY KEY (chat_id, command_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chat_command_levels;
ALTER TABLE chat_members DROP COLUMN IF EXISTS level;
-- +goose StatementEnd
