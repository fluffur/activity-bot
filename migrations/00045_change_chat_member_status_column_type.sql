-- +goose Up
ALTER TABLE chat_members
    ADD COLUMN status_new smallint NOT NULL DEFAULT 0;

UPDATE chat_members
SET status_new = CASE status
                     WHEN 'creator' THEN 5
                     WHEN 'administrator' THEN 3
                     WHEN 'member' THEN 0
                     ELSE 0
    END;

ALTER TABLE chat_members
    DROP COLUMN status;

ALTER TABLE chat_members
    RENAME COLUMN status_new TO status;

-- +goose Down
ALTER TABLE chat_members
    ADD COLUMN status_old text NOT NULL DEFAULT 'member';

UPDATE chat_members
SET status_old = CASE status
                     WHEN 5 THEN 'creator'
                     WHEN 4 THEN 'administrator'
                     WHEN 3 THEN 'administrator'
                     WHEN 2 THEN 'member'
                     WHEN 1 THEN 'member'
                     WHEN 0 THEN 'member'
                     ELSE 'member'
    END;

ALTER TABLE chat_members
    DROP COLUMN status;

ALTER TABLE chat_members
    RENAME COLUMN status_old TO status;