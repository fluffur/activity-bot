-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN gender TEXT NOT NULL DEFAULT 'unknown';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN gender;
-- +goose StatementEnd
