-- +goose Up
CREATE TABLE command_permissions (
    chat_id BIGINT NOT NULL REFERENCES chats(id),
    command_key TEXT NOT NULL,
    required_status SMALLINT NOT NULL,
    PRIMARY KEY (chat_id, command_key)
);

-- +goose Down
DROP TABLE command_permissions;
