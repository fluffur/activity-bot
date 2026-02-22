-- name: EnsureChatExists :one
WITH ins AS (
    INSERT INTO chats (id, norm_warn, newbie_threshold_days)
        VALUES ($1, $2, $3)
        ON CONFLICT (id) DO NOTHING
        RETURNING *)
SELECT *
FROM ins
UNION ALL
SELECT *
FROM chats
WHERE id = $1
LIMIT 1;

-- name: GetOrCreateChat :one
INSERT INTO chats(id, norm_warn)
VALUES ($1, $2)
ON CONFLICT(id) DO UPDATE SET norm = chats.norm
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