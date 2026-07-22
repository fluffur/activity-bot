-- +goose Up
CREATE TABLE crocodile_words
(
    id           BIGSERIAL PRIMARY KEY,
    word         TEXT      NOT NULL UNIQUE,
    category     TEXT      NOT NULL DEFAULT 'general',
    difficulty   SMALLINT  NOT NULL DEFAULT 1,
    used_count   INT       NOT NULL DEFAULT 0,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP
);

CREATE INDEX idx_crocodile_words_random
    ON crocodile_words (used_count);


-- +goose Down

DROP TABLE crocodile_words;