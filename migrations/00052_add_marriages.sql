-- +goose Up
CREATE TYPE marriage_request_status AS ENUM (
    'pending',
    'accepted',
    'rejected',
    'cancelled'
    );

CREATE TABLE marriages
(
    id          BIGSERIAL PRIMARY KEY,
    chat_id     BIGINT      NOT NULL,
    user1_id    BIGINT      NOT NULL,
    user2_id    BIGINT      NOT NULL,
    married_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    divorced_at TIMESTAMPTZ NULL,

    FOREIGN KEY (chat_id, user1_id) REFERENCES chat_members (chat_id, user_id),
    FOREIGN KEY (chat_id, user2_id) REFERENCES chat_members (chat_id, user_id),

    CHECK (user1_id <= user2_id)
);

CREATE UNIQUE INDEX marriages_active_pair_uniq
    ON marriages (chat_id, user1_id, user2_id)
    WHERE divorced_at IS NULL;

CREATE TABLE marriage_requests
(
    id           BIGSERIAL PRIMARY KEY,
    chat_id      BIGINT                  NOT NULL,
    from_user_id BIGINT                  NOT NULL,
    to_user_id   BIGINT                  NOT NULL,
    status       marriage_request_status NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ             NOT NULL DEFAULT now(),
    responded_at TIMESTAMPTZ             NULL,
    FOREIGN KEY (chat_id, from_user_id) REFERENCES chat_members (chat_id, user_id),
    FOREIGN KEY (chat_id, to_user_id) REFERENCES chat_members (chat_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS marriage_requests;
DROP TABLE IF EXISTS marriages;
DROP TYPE IF EXISTS marriage_request_status;