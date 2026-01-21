-- name: AddExemptRequest :exec
INSERT INTO exempt_requests(chat_id, user_id, exempt_until, message_id)
VALUES ($1, $2, $3, $4);

-- name: ApproveExemptRequest :exec
UPDATE exempt_requests
SET status = 'approved'
WHERE chat_id = $1
  AND user_id = $2
  AND message_id = $3;

-- name: RejectExemptRequest :exec
UPDATE exempt_requests
SET status = 'rejected'
WHERE chat_id = $1
  AND user_id = $2
  AND message_id = $3;

-- name: GetExemptRequest :one
SELECT *
FROM exempt_requests
WHERE chat_id = $1
  AND user_id = $2
  AND message_id = $3;