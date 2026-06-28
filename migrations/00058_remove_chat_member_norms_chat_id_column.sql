-- +goose Up
-- +goose StatementBegin

DROP TABLE IF EXISTS chat_member_norms;

CREATE TABLE chat_member_norms
(
    norm_id BIGINT NOT NULL
        REFERENCES chat_norms(id)
            ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    PRIMARY KEY (norm_id, user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS chat_member_norms;

CREATE TABLE chat_member_norms
(
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    norm_id BIGINT NOT NULL
        REFERENCES chat_norms(id)
            ON DELETE CASCADE,
    PRIMARY KEY (chat_id, user_id, norm_id),
    FOREIGN KEY (chat_id, user_id)
        REFERENCES chat_members(chat_id, user_id)
        ON DELETE CASCADE
);

-- +goose StatementEndN