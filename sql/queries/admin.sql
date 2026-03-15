-- name: AddChatAdmin :exec
UPDATE chat_members
SET status = 'administrator'
WHERE chat_id = $1
  AND user_id = $2;

-- name: RemoveChatAdmin :exec
UPDATE chat_members
SET status = 'member'
WHERE chat_id = $1
  AND user_id = $2;

-- name: GetChatAdmins :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1
  AND cm.status IN ('administrator', 'creator')
ORDER BY cm.joined_at;

-- name: IsChatAdmin :one
SELECT EXISTS(SELECT 1
              FROM chat_members
              WHERE chat_id = $1
                AND user_id = $2
                AND status IN ('administrator', 'creator'));

-- name: IsChatCreator :one
SELECT EXISTS(SELECT 1
              FROM chat_members
              WHERE chat_id = $1
                AND user_id = $2
                AND status = 'creator');


-- name: GetChatMemberStatus :one
SELECT status
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;

-- name: GetMembersWithExpiredMute :many
SELECT sqlc.embed(cm)
FROM moderation_actions ma
         JOIN chat_members cm ON ma.chat_id = cm.chat_id AND ma.user_id = cm.user_id
WHERE ma.type = 'mute'
  AND ma.revoked_at IS NULL
  AND ma.expires_at <= NOW();

