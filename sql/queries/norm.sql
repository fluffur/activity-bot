-- name: CreateNorm :exec
INSERT INTO chat_norms(chat_id, name, value)
VALUES ($1, $2, @value)
ON CONFLICT (chat_id, name) DO UPDATE SET value = @value;

-- name: GetNorm :one
SELECT *
FROM chat_norms
WHERE chat_id = $1
  AND name ILIKE '%' || @name::text || '%'
LIMIT 1;

-- name: ListNorms :many
SELECT *
FROM chat_norms
WHERE chat_id = $1;

