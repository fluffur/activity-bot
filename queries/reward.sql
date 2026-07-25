-- name: AddReward :exec
INSERT INTO rewards(chat_id, user_id, author_id, rank, reason)
VALUES ($1, $2, $3, $4, $5);

-- name: RemoveReward :exec
DELETE
FROM rewards
WHERE id = $1;

-- name: ListUserRewards :many
SELECT *
FROM rewards
WHERE chat_id = $1
  AND user_id = $2
ORDER BY created_at;

