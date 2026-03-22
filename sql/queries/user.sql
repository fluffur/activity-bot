-- name: EnsureUserExists :one
INSERT INTO users(id, username, first_name, last_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET username   = $2,
                               first_name = $3,
                               last_name  = $4
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
INSERT INTO users(id, username, first_name, last_name)
SELECT unnest(@ids::bigint[]),
       unnest(@usernames::text[]),
       unnest(@first_names::text[]),
       unnest(@last_names::text[])
ON CONFLICT (id) DO UPDATE SET username   = EXCLUDED.username,
                               first_name = EXCLUDED.first_name,
                               last_name  = EXCLUDED.last_name;

-- name: GetUsersByCustomTitle :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = @chat_id
  AND (
    (length(@tag::text) < 2 AND lower(cm.tag::text) = lower(@tag::text))
        OR
    (length(@tag::text) >= 2 AND cm.tag ILIKE @tag::text || '%')
    )
LIMIT 1;

-- name: SetUserGender :exec
UPDATE users SET gender = $2 WHERE id = $1;

-- name: SetUserEmoji :exec
UPDATE users SET emoji = $2 WHERE id = $1;

-- name: SetUserCustomEmojiID :exec
UPDATE users SET custom_emoji_id = $2 WHERE id = $1;