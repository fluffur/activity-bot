-- name: CreateMessage :one
INSERT INTO messages(chat_id, user_id, created_at, deleted_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: MessageReport :many
WITH real_messages AS (
    SELECT
        m.chat_id,
        m.user_id,
        COUNT(*) AS cnt
    FROM messages m
    WHERE m.chat_id = @chat_id
      AND (
        @from_date::timestamptz IS NULL
            OR (m.created_at >= @from_date AND m.created_at < @to_date)
        )
    GROUP BY m.chat_id, m.user_id
),
     imported_messages AS (
         SELECT
             ia.chat_id,
             ia.user_id,
             SUM(ia.messages_count) AS cnt
         FROM imported_activity ia
         WHERE ia.chat_id = @chat_id
           AND (
             @from_date::timestamptz IS NULL
                 OR (ia.period_start >= @from_date AND ia.period_end <= @to_date)
             )
         GROUP BY ia.chat_id, ia.user_id
     )
SELECT
    cm.user_id,
    u.username,
    u.first_name,
    u.last_name,

    COALESCE(r.cnt, 0) + COALESCE(i.cnt, 0) AS messages_count,

    c.weekly_norm,
    (COALESCE(r.cnt, 0) + COALESCE(i.cnt, 0) >= c.weekly_norm) AS norm_done,

    cm.joined_at,
    c.newbie_threshold_days,
    cm.role,
    cm.custom_title
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN real_messages r
                   ON r.chat_id = cm.chat_id
                       AND r.user_id = cm.user_id
         LEFT JOIN imported_messages i
                   ON i.chat_id = cm.chat_id
                       AND i.user_id = cm.user_id
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND (cm.exempt_until IS NULL OR cm.exempt_until < now())
ORDER BY messages_count DESC;


-- name: ImportActivity :exec
INSERT INTO imported_activity (
    chat_id,
    user_id,
    period_start,
    period_end,
    messages_count
)
VALUES (
           @chat_id,
           @user_id,
           @period_start,
           @period_end,
           @messages_count
       )
ON CONFLICT (chat_id, user_id, period_start, period_end)
    DO UPDATE SET messages_count = EXCLUDED.messages_count;

-- name: ImportActivityBulk :exec
INSERT INTO imported_activity (
    chat_id,
    user_id,
    period_start,
    period_end,
    messages_count
)
SELECT
    @chat_id,
    UNNEST(@user_ids::BIGINT[]),
    @period_start,
    @period_end,
    UNNEST(@messages_counts::INT[])
ON CONFLICT (chat_id, user_id, period_start, period_end)
    DO UPDATE
    SET messages_count = EXCLUDED.messages_count;
