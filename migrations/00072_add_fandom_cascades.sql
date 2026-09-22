-- +goose Up

ALTER TABLE role_categories
DROP CONSTRAINT role_categories_fandom_id_fkey,
    ADD CONSTRAINT role_categories_fandom_id_fkey
        FOREIGN KEY (fandom_id)
        REFERENCES fandoms (id)
        ON DELETE CASCADE;

ALTER TABLE roles
DROP CONSTRAINT roles_category_id_fkey,
    ADD CONSTRAINT roles_category_id_fkey
        FOREIGN KEY (category_id)
        REFERENCES role_categories (id)
        ON DELETE CASCADE;

ALTER TABLE role_aliases
DROP CONSTRAINT role_aliases_role_id_fkey,
    ADD CONSTRAINT role_aliases_role_id_fkey
        FOREIGN KEY (role_id)
        REFERENCES roles (id)
        ON DELETE CASCADE;

ALTER TABLE role_reservations
DROP CONSTRAINT role_reservations_role_id_fkey,
    ADD CONSTRAINT role_reservations_role_id_fkey
        FOREIGN KEY (role_id)
        REFERENCES roles (id)
        ON DELETE CASCADE;


-- +goose Down

ALTER TABLE role_reservations
DROP CONSTRAINT role_reservations_role_id_fkey,
    ADD CONSTRAINT role_reservations_role_id_fkey
        FOREIGN KEY (role_id)
        REFERENCES roles (id);

ALTER TABLE role_aliases
DROP CONSTRAINT role_aliases_role_id_fkey,
    ADD CONSTRAINT role_aliases_role_id_fkey
        FOREIGN KEY (role_id)
        REFERENCES roles (id);

ALTER TABLE roles
DROP CONSTRAINT roles_category_id_fkey,
    ADD CONSTRAINT roles_category_id_fkey
        FOREIGN KEY (category_id)
        REFERENCES role_categories (id);

ALTER TABLE role_categories
DROP CONSTRAINT role_categories_fandom_id_fkey,
    ADD CONSTRAINT role_categories_fandom_id_fkey
        FOREIGN KEY (fandom_id)
        REFERENCES fandoms (id);