-- name: CreateUser :exec
INSERT INTO users(id, username, first_name, last_name, created_at, gender, emoji_json, is_bot)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateUser :exec
UPDATE users
SET username   = $1,
    first_name = $2,
    last_name  = $3
WHERE id = $4;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE LOWER(username) = LOWER($1);


-- name: UpsertUsers :exec
INSERT INTO users(id, username, first_name, last_name, is_bot, emoji)
SELECT unnest(@ids::bigint[]),
       unnest(@usernames::text[]),
       unnest(@first_names::text[]),
       unnest(@last_names::text[]),
       unnest(@is_bots::boolean[]),
       unnest(@emojis::text[])
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
