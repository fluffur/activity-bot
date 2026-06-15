-- name: EnsureUserExists :one
INSERT INTO users(id, username, first_name, last_name, is_bot)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET username   = $2,
                               first_name = $3,
                               last_name  = $4,
                               is_bot     = $5
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE LOWER(username) = LOWER($1);


-- name: UpsertUsers :exec
INSERT INTO users(id, username, first_name, last_name, is_bot)
SELECT unnest(@ids::bigint[]),
       unnest(@usernames::text[]),
       unnest(@first_names::text[]),
       unnest(@last_names::text[]),
       unnest(@is_bots::boolean[])
ON CONFLICT (id) DO UPDATE SET username   = EXCLUDED.username,
                               first_name = EXCLUDED.first_name,
                               last_name  = EXCLUDED.last_name,
                               is_bot     = EXCLUDED.is_bot;

-- name: SetUserGender :exec
UPDATE users
SET gender = $2
WHERE id = $1;

-- name: SetUserEmoji :exec
UPDATE users
SET emoji = $2
WHERE id = $1;

-- name: SetUserEmojiJson :exec
UPDATE users
SET emoji_json = $2
WHERE id = $1;

-- name: SetUserCustomEmojiID :exec
UPDATE users
SET custom_emoji_id = $2
WHERE id = $1;
