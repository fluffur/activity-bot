-- name: BanUser :exec
INSERT INTO banned_users (
    user_id,
    reason,
    expires_at
)
VALUES ($1, $2, $3)
    ON CONFLICT (user_id)
DO UPDATE SET
    reason = EXCLUDED.reason,
           created_at = now(),
           expires_at = EXCLUDED.expires_at;


-- name: UnbanUser :exec
DELETE FROM banned_users
WHERE user_id = $1;


-- name: IsUserBanned :one
SELECT EXISTS (
    SELECT 1
    FROM banned_users
    WHERE user_id = $1
      AND (expires_at IS NULL OR expires_at > now())
) AS banned;


-- name: GetUserBan :one
SELECT
    user_id,
    reason,
    created_at,
    expires_at
FROM banned_users
WHERE user_id = $1
  AND (expires_at IS NULL OR expires_at > now());


-- name: BanChat :exec
INSERT INTO banned_chats (
    chat_id,
    reason,
    expires_at
)
VALUES ($1, $2, $3)
    ON CONFLICT (chat_id)
DO UPDATE SET
    reason = EXCLUDED.reason,
           created_at = now(),
           expires_at = EXCLUDED.expires_at;


-- name: UnbanChat :exec
DELETE FROM banned_chats
WHERE chat_id = $1;


-- name: IsChatBanned :one
SELECT EXISTS (
    SELECT 1
    FROM banned_chats
    WHERE chat_id = $1
      AND (expires_at IS NULL OR expires_at > now())
) AS banned;


-- name: GetChatBan :one
SELECT
    chat_id,
    reason,
    created_at,
    expires_at
FROM banned_chats
WHERE chat_id = $1
  AND (expires_at IS NULL OR expires_at > now());


-- name: GetBannedUsers :many
SELECT
    user_id,
    reason,
    created_at,
    expires_at
FROM banned_users
ORDER BY created_at DESC;


-- name: GetBannedChats :many
SELECT
    chat_id,
    reason,
    created_at,
    expires_at
FROM banned_chats
ORDER BY created_at DESC;