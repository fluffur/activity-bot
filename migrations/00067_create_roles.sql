-- +goose Up
CREATE TABLE fandoms
(
    id      BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES chats (id),
    name    TEXT   NOT NULL,
    UNIQUE (chat_id, name)
);

CREATE TABLE role_categories
(
    id         BIGSERIAL PRIMARY KEY,
    fandom_id  BIGINT      NOT NULL REFERENCES fandoms (id),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (fandom_id, name)
);

CREATE TABLE roles
(
    id          BIGSERIAL PRIMARY KEY,
    category_id BIGINT      NOT NULL REFERENCES role_categories (id),
    name        TEXT        NOT NULL,
    emoji       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (category_id, name)
);

CREATE TABLE role_aliases
(
    id      BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles (id),
    name    TEXT   NOT NULL,
    UNIQUE (role_id, name)
);

CREATE TABLE role_reservations
(
    id         BIGSERIAL PRIMARY KEY,
    chat_id    BIGINT      NOT NULL REFERENCES chats (id),
    role_id    BIGINT      NOT NULL REFERENCES roles (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (chat_id, role_id)
);

-- +goose Down
DROP TABLE role_reservations;
DROP TABLE role_aliases;
DROP TABLE roles;
DROP TABLE role_categories;
DROP TABLE fandoms;
