-- name: CreateMessage :one
INSERT INTO messages(chat_id, user_id, created_at, message_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: MessageReport :many
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,
       COALESCE(m.messages_count, 0) AS messages_count,
       c.norm_warn,
       c.norm_ban,
       cm.joined_at,
       c.newbie_threshold_days,
       cm.status,
       cm.custom_title
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN (SELECT chat_id, user_id, COUNT(*) AS messages_count
                    FROM messages
                    WHERE (@from_date::timestamptz IS NULL OR created_at >= @from_date)
                      AND (@to_date::timestamptz IS NULL OR created_at < @to_date)
                    GROUP BY chat_id, user_id) m ON m.chat_id = cm.chat_id
    AND m.user_id = cm.user_id
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
ORDER BY messages_count DESC;


-- name: MessageReportOne :one
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,

       COUNT(m.id) FILTER (WHERE m.created_at >= date_trunc('day', now() AT TIME ZONE 'Europe/Moscow') AT TIME ZONE
                                                 'Europe/Moscow')          AS day_count,
       COUNT(m.id) FILTER (WHERE m.created_at >= now() - interval '1 day') AS day_rolling_count,
       COUNT(m.id) FILTER (
           WHERE m.created_at >= (
               (date_trunc('day', t) AT TIME ZONE 'Europe/Moscow')
                   - ((extract(isodow from now() AT TIME ZONE 'Europe/Moscow')::int - c.week_start_day + 7) % 7)
                   * interval '1 day'
               )
           )
        ,

       COALESCE(COUNT(m.id) FILTER (WHERE m.created_at >= now() - interval '7 days'),
                0)::bigint                                                 AS week_rolling_count,
       COALESCE(COUNT(m.id) FILTER (WHERE m.created_at >=
                                          date_trunc('month', t) AT TIME ZONE
                                          'Europe/Moscow'), 0)::bigint     AS month_count,
       COALESCE(COUNT(m.id) FILTER (WHERE m.created_at >= now() - interval '30 days'),
                0)::bigint                                                 AS month_rolling_count,
       COALESCE(COUNT(m.id), 0)::bigint                                    AS all_time_count,

       c.norm_ban,
       c.norm_warn,
       cm.joined_at,
       c.newbie_threshold_days,
       cm.status,
       cm.custom_title,
       cm.rest_until,
       cm.left_at
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         CROSS JOIN (SELECT now() AT TIME ZONE 'Europe/Moscow' AS msk_now) t
         LEFT JOIN messages m
                   ON m.chat_id = cm.chat_id
                       AND m.user_id = cm.user_id

WHERE cm.chat_id = @chat_id
  AND cm.user_id = @user_id
GROUP BY cm.user_id,
         u.username,
         u.first_name,
         u.last_name,
         c.norm_warn,
         c.norm_ban,
         cm.joined_at,
         c.newbie_threshold_days,
         cm.status,
         cm.custom_title,
         cm.rest_until,
         cm.left_at;

-- name: MessageActivityByDay :many
SELECT date_trunc('day', m.created_at AT TIME ZONE 'Europe/Moscow')::date AS day,
       COUNT(m.chat_id)                                                   AS messages_count
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

-- name: MessageActivityByDayAll :many
SELECT date_trunc('day', m.created_at AT TIME ZONE 'Europe/Moscow')::date AS day,
       COUNT(*)                                                           AS messages_count
FROM messages m
WHERE m.chat_id = $1
  AND m.created_at >= COALESCE(@from_date, now() - interval '30 days')
  AND m.created_at <= COALESCE(@to_date, now())
GROUP BY day
ORDER BY day;


-- name: InactiveChatMembers :many
SELECT u.*, cm.custom_title, cm.status, cm.rest_until, MAX(m.created_at)::timestamptz AS last_message_at
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
         LEFT JOIN messages m
                   ON m.user_id = cm.user_id AND m.chat_id = cm.chat_id
WHERE cm.left_at IS NULL
  AND cm.chat_id = $1
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
GROUP BY cm.user_id, u.id, u.first_name, u.last_name, u.username, cm.custom_title, cm.status, cm.rest_until
HAVING MAX(m.created_at) IS NULL
    OR MAX(m.created_at) < NOW() - INTERVAL '1 days'
ORDER BY MAX(m.created_at) NULLS FIRST;
