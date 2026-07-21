-- name: CreateChat :exec
INSERT INTO chats(id,
                  newbie_threshold_days,
                  ai_system_prompt,
                  week_start_day,
                  max_warns,
                  command_prefix,
                  allow_prefixless,
                  mentions_per_message,
                  mention_types,
                  title,
                  tags_enabled,
                  week_start_time,
                  removed_at,
                  emojis_enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: GetChatByID :one
SELECT *
FROM chats
WHERE id = $1
LIMIT 1;

-- name: SetChatMentionTypes :exec
UPDATE chats
SET mention_types = $1
WHERE id = @chat_id;

-- name: SetChatSkipSummonConfirmation :exec
UPDATE chats
SET skip_call_confirmation = $1
WHERE id = @chat_id;


-- name: EnsureChatExists :one
WITH ins AS (
    INSERT
        INTO chats (id, title)
            VALUES ($1, $2)
            ON CONFLICT (id) DO
                UPDATE
                SET title = COALESCE(NULLIF(EXCLUDED.title, ''), chats.title)
            RETURNING *)
SELECT *
FROM ins
UNION ALL
SELECT *
FROM chats
WHERE id = $1
LIMIT 1;

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
  AND (@title::text IS NULL OR c.title ILIKE '%' || @title::text || '%')
ORDER BY (SELECT COUNT(*)
          FROM messages m
          WHERE m.chat_id = c.id) DESC
LIMIT 10;

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


-- name: UpdateChatWeekStartTime :exec
UPDATE chats
SET week_start_time = $1
WHERE id = @chat_id;

-- name: GetChatsWithoutTitle :many
SELECT *
FROM chats
WHERE title = ''
  AND id < 0
  AND removed_at IS NULL;

-- name: GetUserManagedChats :many
SELECT c.*
FROM chats c
         JOIN chat_members cm ON c.id = cm.chat_id
WHERE c.id < 0
  AND cm.user_id = $1
  AND cm.status > 0
  AND title <> ''
  AND (@title::text IS NULL OR c.title LIKE '%' || @title::text || '%');

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

-- name: RemoveChat :exec
UPDATE chats
SET removed_at = $1
WHERE id = $2;

-- name: SetChatEmojisEnabled :exec
UPDATE chats
SET emojis_enabled = $1
WHERE id = $2;

-- name: SetAllowPolygamy :exec
UPDATE chats
SET allow_polygamy = $1
WHERE id = $2;

-- name: SetUsernameChangeNotifyStatus :exec
UPDATE chats
SET username_changed_notify_status = $1
WHERE id = $2;