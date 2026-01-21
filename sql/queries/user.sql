-- name: EnsureUserExists :exec
INSERT INTO users(id, username, first_name, last_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET username   = $2,
                               first_name = $3,
                               last_name  = $4;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;