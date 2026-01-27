-- name: GetMemberCustomTitle :one
SELECT custom_title
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;


-- name: EnsureChatMemberExists :one
INSERT INTO chat_members(chat_id, user_id, role)
VALUES ($1, $2, @role)
ON CONFLICT(chat_id, user_id) DO UPDATE SET role    = EXCLUDED.role,
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
SELECT cm.user_id, cm.custom_title, cm.role, u.first_name, u.last_name, u.username
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

-- name: UpdateChatMemberRole :exec
UPDATE chat_members
SET role = @role
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
INTO chat_members (chat_id, user_id, role)
SELECT chat_upsert.id, user_upsert.id, @role
FROM chat_upsert,
     user_upsert
ON CONFLICT (chat_id, user_id) DO UPDATE SET role    = CASE
                                                           WHEN EXCLUDED.role = 'creator' THEN 'creator'
                                                           WHEN chat_members.role = 'administrator' THEN 'administrator'
                                                           WHEN chat_members.role = 'creator' AND EXCLUDED.role <> 'creator'
                                                               THEN 'creator'
                                                           ELSE EXCLUDED.role
    END,
                                             left_at = NULL
RETURNING *;

-- name: UpsertChatMembersWithRole :exec
INSERT INTO chat_members(chat_id, user_id, custom_title, role)
SELECT @chat_id, UNNEST(@user_ids::BIGINT[]), UNNEST(@custom_titles::TEXT[]), UNNEST(@roles::TEXT[])
ON CONFLICT (chat_id, user_id) DO UPDATE SET custom_title = EXCLUDED.custom_title,
                                             role         = CASE
                                                                WHEN EXCLUDED.role = 'creator' THEN 'creator'
                                                                WHEN chat_members.role = 'administrator'
                                                                    THEN 'administrator'
                                                                ELSE EXCLUDED.role
                                                 END,
                                             left_at      = NULL
;

-- name: MarkChatMembersLeftNotInList :exec
UPDATE chat_members
SET left_at = now()
WHERE chat_id = @chat_id
  AND left_at IS NULL
  AND user_id <> ALL(@user_ids::BIGINT[]);


-- name: MoveChatMembersToOldExcept :exec
UPDATE chat_members cm
SET joined_at = joined_at - ((c.newbie_threshold_days + 1) || ' days')::interval
FROM chats c
WHERE c.id = cm.chat_id
  AND cm.chat_id = $1
  AND cm.user_id <> ALL(@user_ids::BIGINT[]);
