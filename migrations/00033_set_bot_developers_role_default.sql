-- +goose Up
-- +goose StatementBegin
ALTER TABLE bot_developers ALTER COLUMN role SET DEFAULT 'member';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE bot_developers ALTER COLUMN role DROP DEFAULT;
-- +goose StatementEnd
