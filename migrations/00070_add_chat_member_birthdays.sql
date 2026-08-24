-- +goose Up
ALTER TABLE chat_members
    ADD COLUMN birthday DATE;

-- +goose Down
ALTER TABLE chat_members DROP COLUMN birthday;