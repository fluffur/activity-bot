-- +goose Up
-- +goose StatementBegin

INSERT INTO chat_norms (chat_id, name, value)
SELECT id, 'варн', norm_warn
FROM chats
WHERE norm_warn IS NOT NULL AND norm_warn > 0
ON CONFLICT (chat_id, name)
    DO UPDATE SET value = EXCLUDED.value;

INSERT INTO chat_norms (chat_id, name, value)
SELECT id, 'бан', norm_ban
FROM chats
WHERE norm_ban IS NOT NULL AND norm_ban > 0
ON CONFLICT (chat_id, name)
    DO UPDATE SET value = EXCLUDED.value;

ALTER TABLE chats
    DROP CONSTRAINT IF EXISTS chats_weekly_norm_check,
    DROP CONSTRAINT IF EXISTS chats_norm_ban_check,
    DROP COLUMN IF EXISTS norm_warn,
    DROP COLUMN IF EXISTS norm_ban;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE chats
    ADD COLUMN norm_warn integer,
    ADD COLUMN norm_ban integer;

ALTER TABLE chats
    ADD CONSTRAINT chats_weekly_norm_check CHECK ((norm_warn IS NULL) OR (norm_warn > 0)),
    ADD CONSTRAINT chats_norm_ban_check CHECK ((norm_ban IS NULL) OR (norm_ban > 0));

UPDATE chats c
SET norm_warn = cn.value
FROM chat_norms cn
WHERE cn.chat_id = c.id AND cn.name = 'warn';

UPDATE chats c
SET norm_ban = cn.value
FROM chat_norms cn
WHERE cn.chat_id = c.id AND cn.name = 'ban';

DELETE FROM chat_norms
WHERE name IN ('warn', 'ban');

-- +goose StatementEnd