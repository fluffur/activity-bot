-- name: EnsureChatExists :one
INSERT INTO chats(id, weekly_norm)
VALUES ($1, $2)
ON CONFLICT(id) DO UPDATE SET weekly_norm = chats.weekly_norm
RETURNING *;

-- name: GetOrCreateChat :one
INSERT INTO chats(id, weekly_norm)
VALUES ($1, $2)
ON CONFLICT(id) DO UPDATE SET weekly_norm = chats.weekly_norm
RETURNING *;

-- name: UpdateChatNorm :exec
UPDATE chats
SET weekly_norm = $1
WHERE id = $2;
