-- name: UpsertRPCommand :exec
INSERT INTO chat_rp_commands (chat_id, trigger, action, emojis, created_by)
VALUES (@chat_id, @trigger, @action, @emojis, @created_by)
ON CONFLICT (chat_id, trigger) DO UPDATE
    SET action     = EXCLUDED.action,
        emojis     = EXCLUDED.emojis,
        updated_at = now();

-- name: DeleteRPCommand :exec
DELETE
FROM chat_rp_commands
WHERE chat_id = @chat_id
  AND trigger = @trigger;

-- name: GetRPCommandByTrigger :one
SELECT *
FROM chat_rp_commands
WHERE chat_id = @chat_id
  AND trigger = @trigger;

-- name: ListRPCommandsByChat :many
SELECT *
FROM chat_rp_commands
WHERE chat_id = @chat_id
ORDER BY trigger;
