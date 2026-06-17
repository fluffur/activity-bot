-- name: GetChatMember :one
SELECT sqlc.embed(chat_members), sqlc.embed(users), sqlc.embed(chats)
FROM chat_members
         JOIN users ON users.id = user_id
         JOIN chats ON chat_members.chat_id = chats.id
    AND chat_id = $1
    AND user_id = $2
    AND users.is_bot = $3;

-- name: CreateChatMember :exec
INSERT INTO chat_members(chat_id, user_id, tag, status, rest_until, left_at, rest_reason, emoji_json)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
