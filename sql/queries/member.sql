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
    INSERT INTO chats (id, weekly_norm)
        VALUES (@chat_id, @weekly_norm)
        ON CONFLICT (id) DO UPDATE SET weekly_norm = chats.weekly_norm
        RETURNING id),
     user_upsert AS (
         INSERT INTO users (id, username, first_name, last_name)
             VALUES (@user_id, @username, @first_name, @last_name)
             ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username,
                 first_name = EXCLUDED.first_name,
                 last_name = EXCLUDED.last_name
             RETURNING id)
INSERT
INTO chat_members (chat_id, user_id, status)
SELECT chat_upsert.id, user_upsert.id, @status
FROM chat_upsert,
     user_upsert
ON CONFLICT (chat_id, user_id) DO UPDATE SET status  = CASE
                                                           WHEN EXCLUDED.status = 'creator' THEN 'creator'
                                                           WHEN chat_members.status = 'administrator'
                                                               THEN 'administrator'
                                                           WHEN chat_members.status = 'creator' AND EXCLUDED.status <> 'creator'
                                                               THEN 'creator'
                                                           ELSE EXCLUDED.status
    END,
                                             left_at = NULL
RETURNING *;

-- name: UpsertChatMembers :exec
INSERT INTO chat_members(chat_id, user_id, custom_title, status)
SELECT @chat_id, UNNEST(@user_ids::BIGINT[]), UNNEST(@custom_titles::TEXT[]), UNNEST(@statuses::TEXT[])
ON CONFLICT (chat_id, user_id) DO UPDATE SET custom_title = EXCLUDED.custom_title,
                                             status         = CASE
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
