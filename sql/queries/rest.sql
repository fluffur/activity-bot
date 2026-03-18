-- name: GetRestMembers :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1
  AND cm.left_at IS NULL
  AND cm.rest_until IS NOT NULL
  AND cm.rest_until >= now()
ORDER BY cm.rest_until;


-- name: SetMemberRest :exec
UPDATE chat_members
SET rest_until  = $1,
    rest_reason = $2
WHERE chat_id = $3
  AND user_id = $4;


-- name: EndMemberRest :exec
UPDATE chat_members
SET rest_until = null
WHERE user_id = $1
  AND chat_id = $2;


-- name: AddRestRequest :exec
INSERT INTO rest_requests(chat_id, user_id, rest_until, message_id, reason, status)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ApproveRestRequest :exec
UPDATE rest_requests
SET status = 'approved'
WHERE chat_id = $1
  AND user_id = $2
  AND message_id = $3;

-- name: RejectRestRequest :exec
UPDATE rest_requests
SET status = 'rejected'
WHERE chat_id = $1
  AND user_id = $2
  AND message_id = $3;

-- name: GetRestRequest :one
SELECT *
FROM rest_requests
WHERE chat_id = $1
  AND user_id = $2
  AND message_id = $3;

-- name: GetAllActiveRests :many
SELECT chat_id, user_id, rest_until, rest_reason
FROM chat_members
WHERE rest_until IS NOT NULL
  AND rest_until >= now();

-- name: GetUserRestRequests :many
SELECT sqlc.embed(rr), sqlc.embed(u), sqlc.embed(cm)
FROM rest_requests rr
         JOIN users u ON u.id = rr.user_id
         JOIN chat_members cm ON cm.user_id = u.id AND cm.chat_id = rr.chat_id
WHERE rr.user_id = $1
  AND rr.chat_id = $2
ORDER BY rr.requested_at DESC;

-- name: DeleteRestRequest :exec
DELETE
FROM rest_requests
WHERE id = $1;