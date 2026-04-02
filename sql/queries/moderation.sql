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
SELECT sqlc.embed(um), sqlc.embed(cmm), sqlc.embed(u), sqlc.embed(cm), ma.*
FROM moderation_actions ma
         JOIN chat_members cmm ON cmm.user_id = ma.moderator_id AND cmm.chat_id = ma.chat_id
         JOIN users um ON um.id = ma.moderator_id

         JOIN chat_members cm ON cm.user_id = ma.user_id AND cm.chat_id = ma.chat_id
         JOIN users u ON u.id = ma.user_id
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
SELECT sqlc.embed(um) AS moderator_user, sqlc.embed(cmm) AS moderator_chat_member, sqlc.embed(u), sqlc.embed(cm), ma.*
FROM moderation_actions ma
         JOIN chat_members cmm ON cmm.user_id = ma.moderator_id AND cmm.chat_id = ma.chat_id
         JOIN users um ON um.id = ma.moderator_id

         JOIN chat_members cm ON cm.user_id = ma.user_id AND cm.chat_id = ma.chat_id
         JOIN users u ON u.id = ma.user_id
WHERE ma.chat_id = $1
  AND ma.type = 'warn'
ORDER BY ma.created_at;
