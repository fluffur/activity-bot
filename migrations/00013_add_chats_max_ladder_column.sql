-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats
    ADD COLUMN max_ladder INTEGER NOT NULL DEFAULT 0
        CONSTRAINT chats_max_ladder_check CHECK ( max_ladder >= 0 );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats
    DROP COLUMN max_ladder;
-- +goose StatementEnd
