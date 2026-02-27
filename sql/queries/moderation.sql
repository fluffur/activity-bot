-- name: CreateModerationAction :exec
INSERT INTO moderation_actions (type, chat_id, user_id, moderator_id, reason, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetActiveWarnsCount :one
SELECT count(*)
FROM moderation_actions
WHERE chat_id = $1
  AND user_id = $2
  AND type = 'warn';

-- name: GetActiveWarns :many
SELECT *
FROM moderation_actions ma
         JOIN users u ON ma.moderator_id = u.id
WHERE ma.chat_id = $1
  AND ma.user_id = $2
  AND ma.type = 'warn'
ORDER BY ma.created_at;

-- name: ClearWarns :exec
DELETE
FROM moderation_actions
WHERE chat_id = $1
  AND user_id = $2
  AND type = 'warn';

-- name: RemoveLatestWarn :exec
DELETE
FROM moderation_actions
WHERE id = (SELECT ma.id
            FROM moderation_actions ma
            WHERE ma.chat_id = @chat_id
              AND ma.user_id = @user_id
              AND ma.type = 'warn'
            ORDER BY ma.created_at DESC
            LIMIT 1);

-- name: GetChatMaxWarns :one
SELECT max_warns
FROM chats
WHERE id = $1;

-- name: UpdateChatMaxWarns :exec
UPDATE chats
SET max_warns = $1
WHERE id = $2;

-- name: DeleteModerationActionsForUser :exec
DELETE
FROM moderation_actions
WHERE chat_id = $1
  AND user_id = $2;

-- name: GetActiveWarnsByChat :many
SELECT ma.*,
       u.username AS user_username, u.first_name AS user_first_name, u.last_name AS user_last_name, u.gender AS user_gender,
       m.username AS mod_username, m.first_name AS mod_first_name, m.last_name AS mod_last_name, m.gender AS mod_gender
FROM moderation_actions ma
         JOIN users u ON ma.user_id = u.id
         JOIN users m ON ma.moderator_id = m.id
WHERE ma.chat_id = $1
  AND ma.type = 'warn'
ORDER BY ma.created_at;
