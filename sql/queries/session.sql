-- name: UpsertPMSession :exec
INSERT INTO user_pm_sessions (user_id, target_chat_id, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id) DO UPDATE
SET target_chat_id = EXCLUDED.target_chat_id,
    updated_at = EXCLUDED.updated_at;

-- name: GetPMSession :one
SELECT target_chat_id
FROM user_pm_sessions
WHERE user_id = $1;

-- name: GetChatPMSession :one
SELECT sqlc.embed(c)
FROM user_pm_sessions
JOIN chats c ON c.id = target_chat_id
WHERE user_id = $1 LIMIT 1;

-- name: DeletePMSession :exec
DELETE FROM user_pm_sessions
WHERE user_id = $1;
