-- +goose Up
ALTER TABLE role_reservations
    ADD COLUMN user_id BIGINT NOT NULL REFERENCES users (id) DEFAULT 0;

-- +goose Down
ALTER TABLE role_reservations DROP COLUMN user_id;