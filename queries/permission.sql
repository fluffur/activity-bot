-- name: GetCommandPermission :one
SELECT *
FROM command_permissions
WHERE chat_id = $1
  AND command_key = $2;

-- name: GetCommandPermissions :many
SELECT *
FROM command_permissions
WHERE chat_id = $1;

-- name: SetCommandPermission :exec
INSERT INTO command_permissions (chat_id, command_key, required_status)
VALUES ($1, $2, $3)
ON CONFLICT (chat_id, command_key) DO UPDATE
    SET required_status = EXCLUDED.required_status;
