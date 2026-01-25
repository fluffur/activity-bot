-- name: AddChatAdmin :exec
INSERT INTO chat_admins(chat_id, user_id)
VALUES ($1, $2)
ON CONFLICT (chat_id, user_id) DO NOTHING;

-- name: RemoveChatAdmin :exec
DELETE
FROM chat_admins
WHERE chat_id = $1
  AND user_id = $2;

-- name: GetChatAdmins :many
SELECT u.id, u.username, u.first_name, u.last_name, ca.created_at
FROM chat_admins ca
         JOIN users u ON u.id = ca.user_id
WHERE ca.chat_id = $1
ORDER BY ca.created_at;

-- name: IsChatAdmin :one
SELECT EXISTS(SELECT 1
              FROM chat_admins
              WHERE chat_id = $1
                AND user_id = $2);

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

