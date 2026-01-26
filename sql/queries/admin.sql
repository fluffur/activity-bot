-- name: AddChatAdmin :exec
UPDATE chat_members
SET role = 'administrator'
WHERE chat_id = $1 AND user_id = $2;

-- name: RemoveChatAdmin :exec
UPDATE chat_members
SET role = 'member'
WHERE chat_id = $1 AND user_id = $2;

-- name: GetChatAdmins :many
SELECT u.id, u.username, u.first_name, u.last_name, cm.joined_at as created_at
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1 AND cm.role IN ('administrator', 'creator')
ORDER BY cm.joined_at;

-- name: IsChatAdmin :one
SELECT EXISTS(SELECT 1
              FROM chat_members
              WHERE chat_id = $1
                AND user_id = $2
                AND role IN ('administrator', 'creator'));

-- name: IsChatCreator :one
SELECT EXISTS(SELECT 1
              FROM chat_members
              WHERE chat_id = $1
                AND user_id = $2
                AND role = 'creator');


-- name: GetChatMemberRole :one
SELECT role
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;

