-- name: GetDeveloper :one
SELECT * FROM bot_developers WHERE user_id = $1 AND chat_id = $2;

-- name: EnsureDeveloperUser :one
INSERT INTO users (id, first_name, last_name)
VALUES ($1, 'Developer', '')
ON CONFLICT (id) DO UPDATE SET first_name = EXCLUDED.first_name
RETURNING *;

-- name: SetDeveloper :exec
INSERT INTO bot_developers (user_id, chat_id, level)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, chat_id) DO UPDATE SET level = EXCLUDED.level;

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
