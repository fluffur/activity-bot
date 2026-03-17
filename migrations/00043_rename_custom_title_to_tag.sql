-- +goose Up
ALTER TABLE chat_members RENAME COLUMN custom_title TO tag;

-- +goose Down
ALTER TABLE chat_members RENAME COLUMN tag TO custom_title;
