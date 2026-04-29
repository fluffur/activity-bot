-- name: UpsertRPCommand :exec
INSERT INTO chat_rp_commands (chat_id, trigger, trigger_normalized, template, emoji_json, created_by)
VALUES (@chat_id, @trigger, @trigger_normalized, @template, @emoji_json, @created_by)
ON CONFLICT (chat_id, trigger_normalized) DO UPDATE
    SET trigger    = EXCLUDED.trigger,
        template   = EXCLUDED.template,
        emoji_json = EXCLUDED.emoji_json,
        updated_at = now();

-- name: DeleteRPCommand :exec
DELETE
FROM chat_rp_commands
WHERE chat_id = @chat_id
  AND trigger_normalized = @trigger_normalized;

-- name: GetRPCommandByTrigger :one
SELECT *
FROM chat_rp_commands
WHERE chat_id = @chat_id
  AND trigger_normalized = @trigger_normalized;

-- name: ListRPCommandsByChat :many
SELECT *
FROM chat_rp_commands
WHERE chat_id = @chat_id
ORDER BY trigger;
