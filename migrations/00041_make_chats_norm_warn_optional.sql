-- +goose Up

ALTER TABLE chats ALTER COLUMN norm_warn DROP NOT NULL;

UPDATE chats SET norm_warn = NULL WHERE norm_warn = 0;
UPDATE chats SET norm_ban = NULL WHERE norm_ban = 0;

ALTER TABLE chats DROP CONSTRAINT chats_weekly_norm_check;
ALTER TABLE chats DROP CONSTRAINT chats_norm_ban_check;

ALTER TABLE chats
    ADD CONSTRAINT chats_weekly_norm_check
        CHECK (norm_warn IS NULL OR norm_warn > 0);

ALTER TABLE chats
    ADD CONSTRAINT chats_norm_ban_check
        CHECK (norm_ban IS NULL OR norm_ban > 0);


-- +goose Down

ALTER TABLE chats DROP CONSTRAINT chats_weekly_norm_check;
ALTER TABLE chats DROP CONSTRAINT chats_norm_ban_check;

ALTER TABLE chats
    ADD CONSTRAINT chats_weekly_norm_check
        CHECK (norm_warn >= 0);

ALTER TABLE chats
    ADD CONSTRAINT chats_norm_ban_check
        CHECK (norm_ban >= 0);

UPDATE chats SET norm_warn = 0 WHERE norm_warn IS NULL;
UPDATE chats SET norm_ban = 0 WHERE norm_ban IS NULL;

ALTER TABLE chats ALTER COLUMN norm_warn SET NOT NULL;