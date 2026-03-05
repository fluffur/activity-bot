-- +goose Up
-- +goose StatementBegin

ALTER TABLE bot_developers RENAME COLUMN role TO level;

ALTER TABLE bot_developers ALTER COLUMN level DROP DEFAULT;

ALTER TABLE bot_developers
    ALTER COLUMN level TYPE SMALLINT USING
        CASE
            WHEN level = 'creator' THEN 5
            WHEN level = 'admin' THEN 3
            ELSE 0
            END;

ALTER TABLE bot_developers
    ALTER COLUMN level SET DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE bot_developers ALTER COLUMN level DROP DEFAULT;

ALTER TABLE bot_developers
    ALTER COLUMN level TYPE TEXT USING
        CASE
            WHEN level = 5 THEN 'creator'
            WHEN level = 3 THEN 'admin'
            ELSE 'member'
            END;

ALTER TABLE bot_developers
    ALTER COLUMN level SET DEFAULT 'member';

ALTER TABLE bot_developers RENAME COLUMN level TO role;

-- +goose StatementEnd