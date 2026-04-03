-- name: EnsureChatExists :one
WITH ins AS (
    INSERT INTO chats (id, title)
        VALUES ($1, $2)
        ON CONFLICT (id) DO UPDATE
            SET title = COALESCE(NULLIF(EXCLUDED.title, ''), chats.title)
        RETURNING *)
SELECT *
FROM ins
UNION ALL
SELECT *
FROM chats
WHERE id = $1
LIMIT 1;

-- name: GetOrCreateChat :one
INSERT INTO chats(id, title, norm_warn)
VALUES ($1, $2, $3)
ON CONFLICT(id) DO UPDATE SET norm_warn = chats.norm_warn,
                              title     = COALESCE(NULLIF(EXCLUDED.title, ''), chats.title)
RETURNING *;

-- name: UpdateChatWarnNorm :exec
UPDATE chats
SET norm_warn = $1
WHERE id = $2;


-- name: UpdateChatBanNorm :exec
UPDATE chats
SET norm_ban = $1
WHERE id = $2;


-- name: UpdateChatNewbieThreshold :exec
UPDATE chats
SET newbie_threshold_days = $1
WHERE id = $2;

-- name: GetChat :one
SELECT *
FROM chats
WHERE id = $1;

-- name: GetAllChats :many
SELECT c.*
FROM chats c
WHERE c.id < 0
  AND c.title <> ''
  AND c.bot_removed_at IS NULL
ORDER BY (SELECT COUNT(*)
          FROM messages m
          WHERE m.chat_id = c.id) DESC;

-- name: UpdateChatTitle :exec
UPDATE chats
SET title = $1
WHERE id = $2;

-- name: SetChatAISystemPrompt :exec
UPDATE chats
SET ai_system_prompt = $1
WHERE id = @chat_id;

-- name: GetChatMaxLadder :one
SELECT max_ladder
FROM chats
WHERE id = @chat_id
LIMIT 1;

-- name: SetChatMaxLadder :exec
UPDATE chats
SET max_ladder = $1
WHERE id = @chat_id;

-- name: SetChatWelcomeCallMessage :exec
UPDATE chats
SET welcome_call_message = $1
WHERE id = @chat_id;


-- name: UpdateChatCallOnJoin :exec
UPDATE chats
SET call_on_join = $1
WHERE id = @chat_id;

-- name: UpdateChatWeekStartDay :exec
UPDATE chats
SET week_start_day = $1
WHERE id = @chat_id;

-- name: UpdateChatCommandPrefix :exec
UPDATE chats
SET command_prefix = $1
WHERE id = @chat_id;

-- name: UpdateChatAllowPrefixless :exec
UPDATE chats
SET allow_prefixless = $1
WHERE id = @chat_id;

-- name: UpdateChatMentionsPerMessage :exec
UPDATE chats
SET mentions_per_message = $1
WHERE id = @chat_id;

-- name: UpdateChatMentionTypes :exec
UPDATE chats
SET mention_types = $1
WHERE id = @chat_id;

-- name: UpdateChatWeekStartTime :exec
UPDATE chats
SET week_start_time = $1
WHERE id = @chat_id;

-- name: GetAllUserChatsWithoutNorm :many
SELECT c.id,
       c.title,
       c.norm_ban,
       c.norm_warn,
       COUNT(m.id) AS week_count
FROM chats c

         JOIN chat_members cm
              ON cm.chat_id = c.id
                  AND cm.user_id = @user_id
                  AND cm.left_at IS NULL

         LEFT JOIN messages m
                   ON m.chat_id = c.id
                       AND m.user_id = @user_id
                       AND m.created_at >= (
                                               date_trunc('day', now())
                                                   - ((extract(isodow from now())::int - c.week_start_day + 7) % 7) *
                                                     interval '1 day'
                                                   + c.week_start_time::interval
                                               ) - CASE
                                                       WHEN now()::time < c.week_start_time THEN interval '7 days'
                                                       ELSE interval '0 days' END

WHERE c.id < 0
  AND c.title <> ''

GROUP BY c.id, c.title, c.norm_ban, c.norm_warn, c.week_start_time

HAVING COUNT(m.id) < GREATEST(c.norm_ban, c.norm_warn)

ORDER BY week_count;

-- name: GetChatsWithoutTitle :many
SELECT *
FROM chats
WHERE title = ''
  AND id < 0;

-- name: GetUserManagedChats :many
SELECT c.*
FROM chats c
         JOIN chat_members cm ON c.id = cm.chat_id
WHERE c.id < 0
  AND cm.user_id = $1
  AND cm.status > 0
  AND title <> '';

-- name: GetChatsWithEnabledBroadcast :many
SELECT c.*
FROM chats c
WHERE c.id < 0
  AND c.title <> ''
  AND c.broadcast_enabled;

-- name: SetChatBroadcast :exec
UPDATE chats
SET broadcast_enabled = $1
WHERE id = $2;

-- name: GetCommandPermissions :many
SELECT *
FROM command_permissions
WHERE chat_id = $1;

-- name: GetCommandPermission :one
SELECT *
FROM command_permissions
WHERE chat_id = $1
  AND command_key = $2;

-- name: SetCommandPermission :exec
INSERT INTO command_permissions (chat_id, command_key, required_status)
VALUES ($1, $2, $3)
ON CONFLICT (chat_id, command_key) DO UPDATE
    SET required_status = EXCLUDED.required_status;

-- name: SetChatDeletedBotAt :exec
UPDATE chats
SET bot_removed_at = $1
WHERE id = $2;