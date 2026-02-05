-- name: CreateMessage :one
INSERT INTO messages(chat_id, user_id, created_at, deleted_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: MessageReport :many
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,
       COUNT(m.chat_id)                    AS messages_count,
       c.weekly_norm,
       (COUNT(m.chat_id) >= c.weekly_norm) AS norm_done,
       cm.joined_at,
       c.newbie_threshold_days,
       cm.status,
       cm.custom_title
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN messages m
                   ON m.chat_id = cm.chat_id
                       AND m.user_id = cm.user_id
                       AND (
                          @from_date::timestamptz IS NULL
                              OR (m.created_at >= @from_date AND m.created_at < @to_date)
                          )
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
GROUP BY cm.user_id, u.username, u.first_name, u.last_name,
         c.weekly_norm, cm.joined_at, c.newbie_threshold_days,
         cm.status, cm.custom_title
ORDER BY messages_count DESC;


-- name: MessageReportOne :one
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,

       COALESCE(COUNT(m.chat_id) FILTER (WHERE m.created_at >= date_trunc('day', now())), 0)::bigint   AS day_count,
       COALESCE(COUNT(m.chat_id) FILTER (WHERE m.created_at >= now() - interval '1 day'),
                0)::bigint                                                                             AS day_rolling_count,
       COALESCE(COUNT(m.chat_id) FILTER (WHERE m.created_at >= date_trunc('week', now())), 0)::bigint  AS week_count,
       COALESCE(COUNT(m.chat_id) FILTER (WHERE m.created_at >= now() - interval '7 days'),
                0)::bigint                                                                             AS week_rolling_count,
       COALESCE(COUNT(m.chat_id) FILTER (WHERE m.created_at >= date_trunc('month', now())), 0)::bigint AS month_count,
       COALESCE(COUNT(m.chat_id) FILTER (WHERE m.created_at >= now() - interval '30 days'),
                0)::bigint                                                                             AS month_rolling_count,
       COALESCE(COUNT(m.chat_id), 0)::bigint                                                           AS all_time_count,

       c.weekly_norm,
       cm.joined_at,
       c.newbie_threshold_days,
       cm.status,
       cm.custom_title,
       cm.rest_until
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN messages m
                   ON m.chat_id = cm.chat_id
                       AND m.user_id = cm.user_id

WHERE cm.chat_id = @chat_id
  AND cm.user_id = @user_id
GROUP BY cm.user_id,
         u.username,
         u.first_name,
         u.last_name,
         c.weekly_norm,
         cm.joined_at,
         c.newbie_threshold_days,
         cm.status,
         cm.custom_title,
         cm.rest_until;

-- name: MessageActivityByDay :many
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

-- name: InactiveChatMembers :many
SELECT u.*, cm.custom_title, cm.status, cm.rest_until, MAX(m.created_at)::date AS last_message_at
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
         LEFT JOIN messages m
                   ON m.user_id = cm.user_id AND m.chat_id = cm.chat_id
WHERE cm.left_at IS NULL
  AND cm.chat_id = $1
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
GROUP BY cm.user_id, u.id, u.first_name, u.last_name, u.username, cm.custom_title, cm.status, cm.rest_until
HAVING MAX(m.created_at) IS NULL
    OR MAX(m.created_at) < NOW() - INTERVAL '1 days';
