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
SELECT *
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
WHERE cm.custom_title ILIKE '%' || @custom_title || '%'
  AND cm.chat_id = @chat_id
LIMIT 10;

-- name: SetUserGender :exec
UPDATE users SET gender = $2 WHERE id = $1;