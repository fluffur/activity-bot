package redis

import (
	"activity-bot/internal/ladder"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ladderLuaScript = `
-- KEYS[1] = ladder key
-- ARGV[1] = userID
-- ARGV[2] = ttl (seconds)

local value = redis.call("GET", KEYS[1])
local userId = ARGV[1]
local ttl = tonumber(ARGV[2])

if not value then
  redis.call("SET", KEYS[1], userId .. ":1", "EX", ttl)
  return {1, 0} -- count, sameUser=false
end

local lastUser, count = value:match("([^:]+):([^:]+)")
count = tonumber(count)

if lastUser == userId then
  count = count + 1
else
  count = 1
end

redis.call("SET", KEYS[1], userId .. ":" .. count, "EX", ttl)

return {count, lastUser == userId and 1 or 0}
`

type LadderRepository struct {
	client *redis.Client
}

func NewLadderRepository(client *redis.Client) ladder.Repository {
	return &LadderRepository{client}
}

func (r *LadderRepository) Inc(
	ctx context.Context,
	chatID, userID int64,
	ttl time.Duration,
) (count int64, sameUser bool, err error) {

	key := ladderKey(chatID)

	res, err := r.client.Eval(ctx, ladderLuaScript, []string{key},
		userID,
		int64(ttl.Seconds()),
	).Result()

	if err != nil {
		return 0, false, err
	}

	values := res.([]interface{})
	count = values[0].(int64)
	sameUser = values[1].(int64) == 1

	return count, sameUser, nil
}

func (r *LadderRepository) Reset(ctx context.Context, chatID int64) error {
	return r.client.Del(ctx, ladderKey(chatID)).Err()
}

func ladderKey(chatID int64) string {
	return fmt.Sprintf("ladder:%d", chatID)
}
