-- name: GetDeveloper :one
SELECT * FROM bot_developers WHERE user_id = $1;

-- name: SetDeveloper :exec
INSERT INTO bot_developers (user_id, role)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: RemoveDeveloper :exec
DELETE FROM bot_developers WHERE user_id = $1;

-- name: IsDeveloper :one
SELECT EXISTS(SELECT 1 FROM bot_developers WHERE user_id = $1);

-- name: GetAllDevelopers :many
SELECT bd.*, u.username, u.first_name, u.last_name
FROM bot_developers bd
JOIN users u ON bd.user_id = u.id
ORDER BY bd.user_id;

-- name: GetDevelopersCount :one
SELECT COUNT(*) FROM bot_developers;
