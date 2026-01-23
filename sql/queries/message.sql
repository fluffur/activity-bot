-- name: CreateMessage :one
INSERT INTO messages(chat_id, user_id, created_at, deleted_at)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: WeeklyMessageReport :many
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,
       COUNT(m.chat_id)                    AS messages_count,
       c.weekly_norm,
       (COUNT(m.chat_id) >= c.weekly_norm) AS norm_done
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
