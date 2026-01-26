-- +goose Up
-- +goose StatementBegin

INSERT INTO chat_members (chat_id, user_id, role)
SELECT chat_id, user_id, 'administrator'
FROM chat_admins
ON CONFLICT (chat_id, user_id) DO UPDATE SET role = 'administrator';

ALTER TABLE chat_members ALTER COLUMN role SET DEFAULT 'member';

UPDATE chat_members
SET role = 'member' 
WHERE role = 'administrator' 
AND NOT EXISTS (
    SELECT 1 FROM chat_admins 
    WHERE chat_admins.chat_id = chat_members.chat_id 
    AND chat_admins.user_id = chat_members.user_id
);

DROP TABLE chat_admins;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

CREATE TABLE chat_admins
(
    chat_id    BIGINT      NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (chat_id, user_id)
);

INSERT INTO chat_admins (chat_id, user_id)
SELECT chat_id, user_id
FROM chat_members
WHERE role IN ('administrator', 'creator');

ALTER TABLE chat_members ALTER COLUMN role SET DEFAULT 'administrator';

-- +goose StatementEnd
