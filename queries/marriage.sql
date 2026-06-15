-- name: CreateMarriageRequest :one
INSERT INTO marriage_requests (chat_id,
                               from_user_id,
                               to_user_id,
                               status,
                               created_at)
VALUES ($1, $2, $3, 'pending', now())
RETURNING *;

-- name: GetActiveMarriageRequest :one
SELECT *
FROM marriage_requests
WHERE chat_id = $1
  AND from_user_id = $2
  AND to_user_id = $3
  AND status = 'pending'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetIncomingMarriageRequests :many
SELECT *
FROM marriage_requests
WHERE chat_id = $1
  AND to_user_id = $2
  AND status = 'pending'
ORDER BY created_at DESC;

-- name: UpdateMarriageRequestStatus :exec
UPDATE marriage_requests
SET status       = $3,
    responded_at = now()
WHERE id = $1
  AND chat_id = $2;

-- name: CreateMarriage :one
INSERT INTO marriages (chat_id,
                       user1_id,
                       user2_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetActiveMarriage :one
SELECT sqlc.embed(m), sqlc.embed(cm1), sqlc.embed(u1), sqlc.embed(cm2), sqlc.embed(u2)
FROM marriages m
         JOIN chat_members cm1 ON cm1.chat_id = m.chat_id AND cm1.user_id = m.user1_id
         JOIN users u1 ON u1.id = cm1.user_id
         JOIN chat_members cm2 ON cm2.chat_id = m.chat_id AND cm2.user_id = m.user2_id
         JOIN users u2 ON u2.id = cm2.user_id
WHERE m.chat_id = $1
  AND m.divorced_at IS NULL
  AND (m.user1_id = $2 OR m.user2_id = $2)
LIMIT 1;

-- name: GetMarriageBetweenUsers :one
SELECT sqlc.embed(m), sqlc.embed(cm1), sqlc.embed(u1), sqlc.embed(cm2), sqlc.embed(u2)
FROM marriages m
         JOIN chat_members cm1 ON cm1.chat_id = m.chat_id AND cm1.user_id = m.user1_id
         JOIN users u1 ON u1.id = cm1.user_id
         JOIN chat_members cm2 ON cm2.chat_id = m.chat_id AND cm2.user_id = m.user2_id
         JOIN users u2 ON u2.id = cm2.user_id
WHERE m.chat_id = $1
  AND m.user1_id = LEAST($2, $3)
  AND m.user2_id = GREATEST($2, $3)
  AND m.divorced_at IS NULL;

-- name: DivorceMarriage :exec
UPDATE marriages
SET divorced_at = now()
WHERE id = $1
  AND chat_id = $2
  AND divorced_at IS NULL;

-- name: ListActiveMarriages :many
SELECT sqlc.embed(m), sqlc.embed(cm1), sqlc.embed(u1), sqlc.embed(cm2), sqlc.embed(u2)
FROM marriages m
         JOIN chat_members cm1 ON cm1.chat_id = m.chat_id AND cm1.user_id = m.user1_id
         JOIN users u1 ON u1.id = cm1.user_id
         JOIN chat_members cm2 ON cm2.chat_id = m.chat_id AND cm2.user_id = m.user2_id
         JOIN users u2 ON u2.id = cm2.user_id
WHERE m.chat_id = $1
  AND m.divorced_at IS NULL
ORDER BY m.married_at DESC;