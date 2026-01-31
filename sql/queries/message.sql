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
       cm.role,
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
  AND (cm.exempt_until IS NULL OR cm.exempt_until < now())
GROUP BY cm.user_id, u.username, u.first_name, u.last_name,
         c.weekly_norm, cm.joined_at, c.newbie_threshold_days,
         cm.role, cm.custom_title
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
