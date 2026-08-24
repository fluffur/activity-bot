-- +goose Up
ALTER TABLE users
    ADD COLUMN birthday DATE;

-- +goose Down
ALTER TABLE users DROP COLUMN birthday;