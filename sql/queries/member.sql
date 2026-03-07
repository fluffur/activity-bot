-- name: GetMemberCustomTitle :one
SELECT custom_title
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;


-- name: EnsureChatMemberExists :one
INSERT INTO chat_members(chat_id, user_id, status)
VALUES ($1, $2, @status)
ON CONFLICT(chat_id, user_id) DO UPDATE SET status  = EXCLUDED.status,
                                            left_at = NULL
RETURNING *;


-- name: GetChatMember :one
SELECT *
FROM chat_members
         JOIN users ON users.id = user_id
WHERE left_at IS NULL
  AND chat_id = $1
  AND user_id = $2
;

-- name: GetChatMembers :many
SELECT *
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL;

-- name: GetChatMembersWithTitles :many
SELECT cm.user_id, cm.custom_title, cm.status, u.first_name, u.last_name, u.username
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND cm.custom_title IS NOT NULL
  AND cm.custom_title <> '';

-- name: GetAnyChatMembersWithTitles :many
SELECT cm.user_id, cm.custom_title, cm.status, u.first_name, u.last_name, u.username
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
WHERE cm.chat_id = @chat_id
  AND cm.custom_title IS NOT NULL
  AND cm.custom_title <> ''
  AND cm.left_at IS NULL
;

-- name: UpdateChatMemberTitle :exec
UPDATE chat_members
SET custom_title = @custom_title
WHERE chat_id = @chat_id
  AND user_id = @user_id;

-- name: DeleteChatMember :exec
UPDATE chat_members
SET left_at = now()
WHERE chat_id = @chat_id
  AND user_id = @user_id
  AND left_at IS NULL;

-- name: UpdateMemberStatus :exec
UPDATE chat_members
SET status = @status
WHERE chat_id = @chat_id
  AND user_id = @user_id;

-- name: EnsureMemberFull :one
WITH chat_upsert AS (
    INSERT INTO chats (id, norm_warn)
        VALUES (@chat_id, @norm_warn)
        ON CONFLICT (id) DO UPDATE
            SET norm_warn = chats.norm_warn
        RETURNING id),
     user_upsert AS (
         INSERT INTO users (id, username, first_name, last_name)
             VALUES (@user_id, @username, @first_name, @last_name)
             ON CONFLICT (id) DO UPDATE
                 SET username = EXCLUDED.username,
                     first_name = EXCLUDED.first_name,
                     last_name = EXCLUDED.last_name
             RETURNING id)
INSERT
INTO chat_members (chat_id, user_id, custom_title)
SELECT chat_upsert.id,
       user_upsert.id,
       @custom_title
FROM chat_upsert,
     user_upsert
ON CONFLICT (chat_id, user_id) DO UPDATE
    SET custom_title = CASE
                           WHEN @custom_title IS NOT NULL AND @custom_title <> ''
                               THEN @custom_title
                           ELSE chat_members.custom_title
        END,
        left_at      = NULL
RETURNING *;

-- name: UpsertChatMembers :exec
INSERT INTO chat_members(chat_id, user_id, custom_title, status)
SELECT @chat_id, UNNEST(@user_ids::BIGINT[]), UNNEST(@custom_titles::TEXT[]), UNNEST(@statuses::TEXT[])
ON CONFLICT (chat_id, user_id) DO UPDATE SET custom_title = CASE
                                                                WHEN EXCLUDED.custom_title <> ''
                                                                    THEN EXCLUDED.custom_title
                                                                ELSE chat_members.custom_title
    END,
                                             status       = CASE
                                                                WHEN EXCLUDED.status = 'creator' THEN 'creator'
                                                                WHEN chat_members.status = 'administrator'
                                                                    THEN 'administrator'
                                                                ELSE EXCLUDED.status
                                                 END,
                                             left_at      = NULL
;

-- name: MarkChatMembersLeftNotInList :exec
UPDATE chat_members
SET left_at = now()
WHERE chat_id = @chat_id
  AND left_at IS NULL
  AND user_id <> ALL (@user_ids::BIGINT[]);


-- name: MoveChatMembersToOldExcept :exec
UPDATE chat_members cm
SET joined_at = joined_at - ((c.newbie_threshold_days + 1) || ' days')::interval
FROM chats c
WHERE c.id = cm.chat_id
  AND cm.chat_id = $1
  AND cm.user_id <> ALL (@user_ids::BIGINT[]);

-- name: MoveChatMembersToNew :exec
UPDATE chat_members cm
SET joined_at = now()
FROM chats c
WHERE c.id = cm.chat_id
  AND cm.chat_id = $1
  AND cm.user_id = ANY (@user_ids::BIGINT[]);

-- name: GetNoNormMembers :many
SELECT cm.*, u.*
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN (SELECT chat_id, user_id, COUNT(*) AS msg_count
                    FROM messages
                    WHERE (messages.created_at >= @from_date OR @from_date::timestamptz IS NULL)
                      AND (messages.created_at < @to_date OR @to_date::timestamptz IS NULL)
                    GROUP BY chat_id, user_id) m ON m.chat_id = cm.chat_id AND m.user_id = cm.user_id

WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
  AND (
    (@mode = 'warn' AND COALESCE(m.msg_count, 0) > c.norm_ban AND COALESCE(m.msg_count, 0) < c.norm_warn)
        OR (@mode = 'ban' AND COALESCE(m.msg_count, 0) < c.norm_ban)
        OR (@mode = 'any' AND COALESCE(m.msg_count, 0) < GREATEST(c.norm_warn, c.norm_ban))
    );