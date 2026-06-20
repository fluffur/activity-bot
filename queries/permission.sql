-- name: GetCommandPermission :one
SELECT *
FROM command_permissions
WHERE chat_id = $1
  AND command_key = $2;
