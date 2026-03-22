-- name: SetChatMemberStatus :exec
UPDATE chat_members
SET status = $1
WHERE chat_id = $2
  AND user_id = $3;

-- name: GetChatAdmins :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1
  AND cm.status > 0
ORDER BY cm.joined_at;

-- name: GetMembersWithExpiredMute :many
SELECT sqlc.embed(cm)
FROM moderation_actions ma
         JOIN chat_members cm ON ma.chat_id = cm.chat_id AND ma.user_id = cm.user_id
WHERE ma.type = 'mute'
  AND ma.revoked_at IS NULL
  AND ma.expires_at <= NOW();

