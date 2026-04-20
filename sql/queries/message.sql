-- name: CreateMessage :one
INSERT INTO messages(chat_id, user_id, created_at, message_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ChatMemberMessageStatsByChat :many
WITH filtered_messages AS (SELECT m.chat_id, m.user_id
                           FROM messages m
                           WHERE m.chat_id = @chat_id
                             AND (@from_date::timestamptz IS NULL OR m.created_at >= @from_date::timestamptz)
                             AND (@to_date::timestamptz IS NULL OR m.created_at < @to_date::timestamptz))
SELECT sqlc.embed(cm),
       sqlc.embed(u),
       COUNT(fm.chat_id) AS messages_count,
       sqlc.embed(c)
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN filtered_messages fm
                   ON fm.chat_id = cm.chat_id
                       AND fm.user_id = cm.user_id
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
GROUP BY cm.chat_id, cm.user_id, u.id, c.id
ORDER BY messages_count DESC;

-- name: ChatMemberMessageStatsByUser :one
WITH user_messages AS (SELECT m.created_at
                       FROM messages m
                       WHERE m.chat_id = @chat_id
                         AND m.user_id = @user_id)
SELECT sqlc.embed(cm),
       sqlc.embed(u),
       sqlc.embed(c),
       COUNT(*) FILTER (WHERE m.created_at >= date_trunc('day', now()))   AS day_count,
       COUNT(*) FILTER (WHERE m.created_at >= now() - interval '1 day')   AS day_rolling_count,
       COUNT(*) FILTER (WHERE m.created_at >= @from_date)                 AS week_count,
       COUNT(*) FILTER (WHERE m.created_at >= now() - interval '7 days')  AS week_rolling_count,
       COUNT(*) FILTER (WHERE m.created_at >= date_trunc('month', now())) AS month_count,
       COUNT(*) FILTER (WHERE m.created_at >= now() - interval '30 days') AS month_rolling_count,
       COUNT(*)                                                           AS all_time_count
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN user_messages m ON TRUE
WHERE cm.chat_id = @chat_id
  AND cm.user_id = @user_id
GROUP BY cm.chat_id, cm.user_id, u.id, c.id;

-- name: UserMessageActivityDaily :many
SELECT date_trunc('day', m.created_at)::date AS day,
       COUNT(m.chat_id)                      AS messages_count
FROM messages m
         JOIN chat_members cm ON cm.chat_id = m.chat_id AND cm.user_id = m.user_id
WHERE m.chat_id = @chat_id
  AND m.user_id = @user_id
  AND m.created_at >= GREATEST(
        now() - interval '30 days',
        cm.joined_at
                      )
GROUP BY day
ORDER BY day;

-- name: ChatMessageActivityDaily :many
SELECT date_trunc('day', m.created_at)::date AS day,
       COUNT(*)                              AS messages_count
FROM messages m
WHERE m.chat_id = $1
  AND m.created_at >= COALESCE(@from_date, now() - interval '30 days')
  AND m.created_at <= COALESCE(@to_date, now())
GROUP BY day
ORDER BY day;

