-- name: GetDeveloper :one
SELECT * FROM bot_developers WHERE user_id = $1 AND chat_id = $2;

-- name: EnsureDeveloperUser :exec
INSERT INTO users (id, first_name, last_name)
VALUES ($1, 'Developer', '')
ON CONFLICT (id) DO NOTHING;

-- name: SetDeveloper :exec
UPDATE chat_members SET status = $1
WHERE chat_id = $2 AND user_id = $3;

-- name: RemoveDeveloper :exec
DELETE FROM bot_developers WHERE user_id = $1 AND chat_id = $2;

-- name: IsDeveloper :one
SELECT EXISTS(SELECT 1 FROM bot_developers WHERE user_id = $1 AND chat_id = $2);

-- name: GetAllDevelopers :many
SELECT bd.*, u.username, u.first_name, u.last_name
FROM bot_developers bd
JOIN users u ON bd.user_id = u.id
WHERE bd.chat_id = $1
ORDER BY bd.user_id;

-- name: GetDevelopersCount :one
SELECT COUNT(*) FROM bot_developers WHERE chat_id = $1;
