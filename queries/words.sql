-- name: InsertCrocodileWord :exec
INSERT INTO crocodile_words (word)
VALUES ($1)
ON CONFLICT(word) DO NOTHING;


-- name: GetRandomCrocodileWord :one
SELECT *
FROM crocodile_words
WHERE last_used_at IS NULL
   OR last_used_at < NOW() - INTERVAL '7 days'
ORDER BY random()
LIMIT 1;


-- name: MarkCrocodileWordUsed :exec
UPDATE crocodile_words
SET
    used_count = used_count + 1,
    last_used_at = NOW()
WHERE id = $1;