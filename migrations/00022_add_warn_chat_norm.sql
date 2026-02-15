-- +goose Up

ALTER TABLE chats
    RENAME COLUMN weekly_norm TO norm_warn;

ALTER TABLE chats
    ADD COLUMN norm_ban INT CHECK (norm_ban >= 0);


-- +goose Down
ALTER TABLE chats
    RENAME COLUMN norm_warn TO weekly_norm;

ALTER TABLE chats
    DROP COLUMN norm_ban;

