-- name: GetMemberCustomTitle :one
SELECT custom_title
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;


-- name: EnsureChatMemberExists :one
INSERT INTO chat_members(chat_id, user_id)
VALUES ($1, $2)
ON CONFLICT(chat_id, user_id) DO UPDATE SET custom_title = chat_members.custom_title
RETURNING *;



-- name: GetChatMember :one
SELECT *
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;



-- name: GetChatMembersWithTitles :many
SELECT cm.user_id, cm.custom_title, u.first_name, u.last_name, u.username
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
WHERE cm.chat_id = @chat_id
  AND cm.custom_title IS NOT NULL
  AND cm.custom_title <> '';

-- name: UpdateChatMemberTitle :exec
UPDATE chat_members
SET custom_title = @custom_title
WHERE chat_id = @chat_id
  AND user_id = @user_id;

-- name: DeleteChatMember :exec
DELETE
FROM chat_members
WHERE chat_id = @chat_id
  AND user_id = @user_id;
