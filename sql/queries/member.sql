-- name: GetMemberCustomTitle :one
SELECT tag
FROM chat_members
WHERE chat_id = $1
  AND user_id = $2;

-- name: GetChatMember :one
SELECT sqlc.embed(chat_members), sqlc.embed(users)
FROM chat_members
         JOIN users ON users.id = user_id
    AND chat_id = $1
    AND user_id = $2;

-- name: GetChatMembers :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
ORDER BY cm.joined_at;

-- name: GetChatMembersWithTitles :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND cm.tag IS NOT NULL
  AND cm.tag <> ''
ORDER BY cm.tag COLLATE "und-x-icu";



-- name: GetAnyChatMembersWithTitles :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
WHERE cm.chat_id = @chat_id
  AND cm.tag IS NOT NULL
  AND cm.tag <> ''
  AND cm.left_at IS NULL;

-- name: UpdateChatMemberTitle :exec
UPDATE chat_members
SET tag = @tag
WHERE chat_id = @chat_id
  AND user_id = @user_id;

-- name: DeleteChatMember :exec
UPDATE chat_members
SET left_at = now()
WHERE chat_id = @chat_id
  AND user_id = @user_id
  AND left_at IS NULL;

-- name: UpdateMemberStatus :exec
UPDATE chat_members
SET status = @status
WHERE chat_id = @chat_id
  AND user_id = @user_id;

-- name: EnsureMemberFull :one
WITH chat_upsert AS (
    INSERT INTO chats (id)
        VALUES (@chat_id)
        ON CONFLICT (id) DO NOTHING
        RETURNING id),
     chat_id_resolve AS (SELECT id
                         FROM chat_upsert
                         UNION ALL
                         SELECT id
                         FROM chats
                         WHERE id = @chat_id
                         LIMIT 1),
     user_upsert AS (
         INSERT INTO users (id, username, first_name, last_name)
             VALUES (@user_id, @username, @first_name, @last_name)
             ON CONFLICT (id) DO UPDATE
                 SET username = EXCLUDED.username,
                     first_name = EXCLUDED.first_name,
                     last_name = EXCLUDED.last_name
             RETURNING id),
     member_upsert AS (
         INSERT INTO chat_members (chat_id, user_id, tag)
             SELECT chat_id_resolve.id,
                    user_upsert.id,
                    @tag
             FROM chat_id_resolve,
                  user_upsert
             ON CONFLICT (chat_id, user_id) DO UPDATE
                 SET tag = CASE
                               WHEN @tag IS NOT NULL AND @tag <> ''
                                   THEN @tag
                               ELSE chat_members.tag
                     END,
                     left_at = NULL
             RETURNING *)
SELECT sqlc.embed(chat_members),
       sqlc.embed(users)
FROM member_upsert cm
         JOIN chat_members ON chat_members.chat_id = cm.chat_id
    AND chat_members.user_id = cm.user_id
         JOIN users ON users.id = cm.user_id;

-- name: ResetChatMemberCreators :exec
UPDATE chat_members
SET status = 0
WHERE chat_id = @chat_id
  AND status = 5;


-- name: UpsertChatMembers :exec
INSERT INTO chat_members(chat_id, user_id, tag, status)
SELECT @chat_id, UNNEST(@user_ids::BIGINT[]), UNNEST(@tags::TEXT[]), UNNEST(@statuses::SMALLINT[])
ON CONFLICT (chat_id, user_id) DO UPDATE SET tag     = CASE
                                                           WHEN EXCLUDED.tag <> ''
                                                               THEN EXCLUDED.tag
                                                           ELSE chat_members.tag
    END,
                                             status  = GREATEST(EXCLUDED.status, chat_members.status),
                                             left_at = NULL
;

-- name: MarkChatMembersLeftNotInList :exec
UPDATE chat_members
SET left_at = now()
WHERE chat_id = @chat_id
  AND left_at IS NULL
  AND user_id <> ALL (@user_ids::BIGINT[]);


-- name: MoveChatMembersToOldExcept :exec
UPDATE chat_members cm
SET joined_at = joined_at - ((c.newbie_threshold_days + 1) || ' days')::interval
FROM chats c
WHERE c.id = cm.chat_id
  AND cm.chat_id = $1
  AND cm.user_id <> ALL (@user_ids::BIGINT[]);

-- name: MoveChatMembersToNew :exec
UPDATE chat_members cm
SET joined_at = now()
FROM chats c
WHERE c.id = cm.chat_id
  AND cm.chat_id = $1
  AND cm.user_id = ANY (@user_ids::BIGINT[]);

-- name: GetNoNormMembers :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN chats c ON c.id = cm.chat_id
         JOIN users u ON u.id = cm.user_id
         LEFT JOIN (SELECT chat_id, user_id, COUNT(*) AS msg_count
                    FROM messages
                    WHERE (messages.created_at >= @from_date OR @from_date::timestamptz IS NULL)
                      AND (messages.created_at < @to_date OR @to_date::timestamptz IS NULL)
                    GROUP BY chat_id, user_id) m ON m.chat_id = cm.chat_id AND m.user_id = cm.user_id

WHERE cm.chat_id = @chat_id
  AND cm.left_at IS NULL
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
  AND (
    (@mode = 'warn' AND (c.norm_ban IS NULL OR COALESCE(m.msg_count, 0) > c.norm_ban) AND
     COALESCE(m.msg_count, 0) < c.norm_warn)
        OR (@mode = 'ban' AND COALESCE(m.msg_count, 0) < c.norm_ban)
        OR (@mode = 'any' AND COALESCE(m.msg_count, 0) < GREATEST(c.norm_warn, c.norm_ban))
    );

-- name: FindChatMembersByTag :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = @chat_id
  AND (
    (length(@tag::text) < 2 AND lower(cm.tag::text) = lower(@tag::text))
        OR
    (length(@tag::text) >= 2 AND cm.tag ILIKE '%' || @tag::text || '%')
    )
ORDER BY cm.left_at IS NOT NULL, cm.left_at;
;

-- name: GetChatMemberByUsername :one
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1
  AND u.username ILIKE $2
LIMIT 1;

-- name: SetChatMemberEmoji :exec
UPDATE chat_members
SET emoji = $1
WHERE user_id = $2
  AND chat_id = $3;

-- name: SetChatMemberEmojiJSON :exec
UPDATE chat_members
SET emoji_json = $1
WHERE user_id = $2
  AND chat_id = $3;


-- name: SetChatMemberStatus :exec
UPDATE chat_members
SET status = $1
WHERE chat_id = $2
  AND user_id = $3;

-- name: GetChatAdmins :many
SELECT sqlc.embed(cm), sqlc.embed(u)
FROM chat_members cm
         JOIN users u ON u.id = cm.user_id
WHERE cm.chat_id = $1
  AND cm.status > 0
ORDER BY cm.joined_at;

-- name: GetMembersWithExpiredMute :many
SELECT sqlc.embed(cm)
FROM moderation_actions ma
         JOIN chat_members cm ON ma.chat_id = cm.chat_id AND ma.user_id = cm.user_id
WHERE ma.type = 'mute'
  AND ma.revoked_at IS NULL
  AND ma.expires_at <= NOW();


-- name: InactiveChatMembers :many
SELECT sqlc.embed(u),
       cm.tag,
       cm.status,
       cm.rest_until,
       MAX(m.created_at)::timestamptz AS last_message_at
FROM chat_members cm
         JOIN users u ON cm.user_id = u.id
         LEFT JOIN messages m
                   ON m.user_id = cm.user_id AND m.chat_id = cm.chat_id
WHERE cm.left_at IS NULL
  AND cm.chat_id = $1
  AND (cm.rest_until IS NULL OR cm.rest_until < now())
GROUP BY cm.user_id, u.id, u.first_name, u.last_name, u.username, cm.tag, cm.status, cm.rest_until
HAVING MAX(m.created_at) IS NULL
    OR MAX(m.created_at) < NOW() - INTERVAL '1 days'
ORDER BY MAX(m.created_at) NULLS FIRST;

-- name: SetChatMemberExcludeFromCall :exec
UPDATE chat_members
SET exclude_from_call = $1
WHERE user_id = $2
  AND chat_id = $3;