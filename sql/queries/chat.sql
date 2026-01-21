-- name: EnsureChatExists :exec
INSERT INTO chats(id, weekly_norm)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: EnsureChatMemberExists :exec
INSERT INTO chat_members(chat_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetOrCreateChat :one
INSERT INTO chats(id, weekly_norm)
VALUES ($1, $2)
ON CONFLICT(id) DO UPDATE SET weekly_norm = chats.weekly_norm
RETURNING *;

-- name: UpdateChatNorm :exec
UPDATE chats
SET weekly_norm = $1
WHERE id = $2;

-- name: ChatExemptUsers :many
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,
       cm.exempt_until
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1
  AND cm.exempt_until IS NOT NULL
  AND cm.exempt_until >= now()
ORDER BY cm.exempt_until ASC;

-- name: WeeklyMessageReport :many
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,
       COUNT(*)                    AS messages_count,
       c.weekly_norm,
       (COUNT(*) >= c.weekly_norm) AS norm_done
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN messages m
                   ON m.chat_id = cm.chat_id
                       AND m.user_id = cm.user_id
                       AND m.created_at >= $1
                       AND m.created_at < $1 + interval '7 days'
WHERE cm.chat_id = $2
  AND (cm.exempt_until IS NULL OR cm.exempt_until < now())
GROUP BY cm.user_id, u.username, u.first_name, u.last_name, c.weekly_norm
ORDER BY messages_count DESC;

-- name: ExemptChatMember :exec
UPDATE chat_members
SET exempt_until = $1
WHERE chat_id = $2
  AND user_id = $3;

-- name: GetChatMember :one
SELECT *
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;

-- name: RemoveChatMemberExempt :exec
UPDATE chat_members
SET exempt_until = null
WHERE user_id = $1
  AND chat_id = $2;

-- name: AddChatAdmin :exec
INSERT INTO chat_admins(chat_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveChatAdmin :exec
DELETE FROM chat_admins
WHERE chat_id = $1
  AND user_id = $2;

-- name: GetChatAdmins :many
SELECT u.id, u.username, u.first_name, u.last_name, ca.created_at
FROM chat_admins ca
         JOIN users u ON u.id = ca.user_id
WHERE ca.chat_id = $1
ORDER BY ca.created_at;

-- name: IsChatAdmin :one
SELECT EXISTS(
    SELECT 1
    FROM chat_admins
    WHERE chat_id = $1
      AND user_id = $2
);
-- name: UpsertChatMembers :exec
INSERT INTO chat_members(chat_id, user_id)
SELECT @chat_id,
       unnest(@user_ids::bigint[])
ON CONFLICT DO NOTHING;
