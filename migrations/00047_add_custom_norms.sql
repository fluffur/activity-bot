-- +goose Up
CREATE TABLE chat_norms
(
    id      BIGSERIAL PRIMARY KEY,
    chat_id BIGINT  NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    name    TEXT    NOT NULL,
    value   INTEGER NOT NULL,
    UNIQUE (chat_id, name)
);

CREATE TABLE chat_member_norms
(
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    norm_id BIGINT NOT NULL REFERENCES chat_norms (id) ON DELETE CASCADE,
    PRIMARY KEY (chat_id, user_id, norm_id),
    FOREIGN KEY (chat_id, user_id) REFERENCES chat_members (chat_id, user_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chat_member_norms;
DROP TABLE chat_norms;
