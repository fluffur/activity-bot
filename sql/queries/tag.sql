-- name: SetChatTagsEnabled :exec
UPDATE chats
SET tags_enabled = $1
WHERE id = $2;

