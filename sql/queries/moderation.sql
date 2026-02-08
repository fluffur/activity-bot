-- name: CreateModerationAction :exec
INSERT INTO moderation_actions (type, chat_id, user_id, mod_id, reason, until_date)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetActiveWarnsCount :one
SELECT count(*) FROM moderation_actions
WHERE chat_id = $1 AND user_id = $2 AND type = 'warn';

-- name: ClearWarns :exec
DELETE FROM moderation_actions 
WHERE chat_id = $1 AND user_id = $2 AND type = 'warn';

-- name: RemoveLatestWarn :exec
DELETE FROM moderation_actions
WHERE id = (
    SELECT ma.id FROM moderation_actions ma
    WHERE ma.chat_id = @chat_id AND ma.user_id = @user_id AND ma.type = 'warn'
    ORDER BY ma.created_at DESC
    LIMIT 1
);

-- name: GetChatMaxWarns :one
SELECT max_warns FROM chats WHERE id = $1;

-- name: UpdateChatMaxWarns :exec
UPDATE chats SET max_warns = $1 WHERE id = $2;

-- name: DeleteModerationActionsForUser :exec
DELETE FROM moderation_actions WHERE chat_id = $1 AND user_id = $2;
