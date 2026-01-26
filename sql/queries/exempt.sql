-- name: ChatExemptUsers :many
SELECT cm.user_id,
       u.username,
       u.first_name,
       u.last_name,
       cm.exempt_until,
       cm.role,
       cm.custom_title
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1
  AND cm.exempt_until IS NOT NULL
  AND cm.exempt_until >= now()
ORDER BY cm.exempt_until;


-- name: ExemptChatMember :exec
UPDATE chat_members
SET exempt_until = $1
WHERE chat_id = $2
  AND user_id = $3;


-- name: RemoveChatMemberExempt :exec
UPDATE chat_members
SET exempt_until = null
WHERE user_id = $1
  AND chat_id = $2;


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