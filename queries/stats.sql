-- name: ChatStats :many
SELECT
    sqlc.embed(cm),
    sqlc.embed(u),
    COUNT(m.chat_id) AS messages_count
FROM chat_members cm
         JOIN users u
              ON u.id = cm.user_id
         LEFT JOIN messages m
                   ON m.chat_id = cm.chat_id
                       AND m.user_id = cm.user_id
                       AND (@from_date::timestamptz IS NULL OR m.created_at >= @from_date::timestamptz)
                       AND (@to_date::timestamptz IS NULL OR m.created_at < @to_date::timestamptz)
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
  AND NOT u.is_bot
GROUP BY cm.chat_id, cm.user_id, u.id
ORDER BY messages_count DESC;