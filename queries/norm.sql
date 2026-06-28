-- name: SetNorm :one
INSERT INTO chat_norms(chat_id, name, value)
VALUES ($1, $2, @value)
ON CONFLICT (chat_id, name) DO UPDATE SET value = @value
RETURNING id;

-- name: GetNorm :one
SELECT *
FROM chat_norms
WHERE chat_id = $1
  AND name ILIKE '%' || @name::text || '%'
LIMIT 1;

-- name: GetNormMembers :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_norms cn
         JOIN chat_member_norms cmn ON cmn.norm_id = cn.id
         JOIN chat_members cm ON cm.user_id = cmn.user_id AND cm.chat_id = cn.chat_id
         JOIN users u ON u.id = cm.user_id
WHERE cn.id = $1;

-- name: ListNorms :many
SELECT *
FROM chat_norms
WHERE chat_id = $1;

-- name: ListNormsWithMembers :many
SELECT sqlc.embed(n),
       cmn.user_id
FROM chat_norms n
         LEFT JOIN chat_member_norms cmn
                   ON cmn.norm_id = n.id
WHERE n.chat_id = $1
ORDER BY n.id;

-- name: DeleteNorm :exec
DELETE
FROM chat_norms
WHERE id = $1;

-- name: AssignNormMembers :exec
INSERT INTO chat_member_norms (norm_id, user_id)
SELECT @norm_id, unnest(@user_ids::bigint[])
ON CONFLICT (norm_id, user_id) DO NOTHING;

-- name: UnassignNormMembers :exec
DELETE
FROM chat_member_norms
WHERE norm_id = @norm_id
  AND user_id = ANY (@user_ids::bigint[]);
