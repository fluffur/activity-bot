-- +goose Up
-- +goose StatementBegin

ALTER TABLE rest_requests
    ADD COLUMN id BIGSERIAL;

ALTER TABLE rest_requests
    ADD COLUMN updated_at timestamptz;

ALTER TABLE rest_requests
    DROP CONSTRAINT exempt_requests_pkey;

ALTER TABLE rest_requests
    ADD CONSTRAINT rest_requests_pkey PRIMARY KEY (id);

ALTER TABLE rest_requests
    ALTER COLUMN message_id DROP NOT NULL;

CREATE UNIQUE INDEX rest_requests_unique
    ON rest_requests (chat_id, user_id, rest_until);

INSERT INTO rest_requests (chat_id, user_id, rest_until, status, reason, requested_at, updated_at)
SELECT cm.chat_id,
       cm.user_id,
       cm.rest_until,
       'approved'::rest_status,
       MAX(cm.rest_reason) AS reason,
       now(),
       now()
FROM chat_members cm
         LEFT JOIN rest_requests rr
                   ON rr.chat_id = cm.chat_id
                       AND rr.user_id = cm.user_id
                       AND rr.rest_until = cm.rest_until
                       AND rr.status = 'approved'
WHERE cm.rest_until IS NOT NULL
  AND rr.id IS NULL
GROUP BY cm.chat_id, cm.user_id, cm.rest_until;

CREATE OR REPLACE FUNCTION rest_requests_set_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_rest_requests_updated_at
    BEFORE INSERT OR UPDATE ON rest_requests
    FOR EACH ROW
EXECUTE FUNCTION rest_requests_set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_rest_requests_updated_at ON rest_requests;
DROP FUNCTION IF EXISTS rest_requests_set_updated_at();

DROP INDEX IF EXISTS rest_requests_unique;

ALTER TABLE rest_requests
    DROP CONSTRAINT rest_requests_pkey;

ALTER TABLE rest_requests
    DROP COLUMN id;

ALTER TABLE rest_requests
    DROP COLUMN updated_at;

ALTER TABLE rest_requests
    ADD CONSTRAINT exempt_requests_pkey
        PRIMARY KEY (chat_id, user_id, message_id);

-- +goose StatementEnd