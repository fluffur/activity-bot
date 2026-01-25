-- name: EnsureChatExists :one
INSERT INTO chats(id, weekly_norm, newbie_threshold_days)
VALUES ($1, $2, $3)
ON CONFLICT(id) DO UPDATE SET weekly_norm = EXCLUDED.weekly_norm,
                               newbie_threshold_days = EXCLUDED.newbie_threshold_days
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

-- name: UpdateChatNewbieThreshold :exec
UPDATE chats
SET newbie_threshold_days = $1
WHERE id = $2;

-- name: GetChat :one
SELECT * FROM chats WHERE id = $1;
