package guard

import (
	"activity-bot/internal/cmd"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Redis    *redis.Client
	Limit    int
	Interval time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, interval time.Duration) cmd.Guard {
	return &RateLimiter{
		Redis:    rdb,
		Limit:    limit,
		Interval: interval,
	}
}

func (r *RateLimiter) Check(ctx *ext.Context, command string) (bool, string) {
	cctx := context.Background()
	chatID := ctx.EffectiveChat.Id
	key := fmt.Sprintf("rate:%d:%s", chatID, command)
	count, err := r.Redis.Get(cctx, key).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, ""
	}

	if count >= r.Limit {
		ttl, err := r.Redis.TTL(cctx, key).Result()
		if err != nil || ttl < 0 {
			ttl = r.Interval
		}
		seconds := int(ttl.Seconds())

		return false, fmt.Sprintf("⚠️ Слишком много запросов! Попробуйте через %d секунд.", seconds)
	}

	pipe := r.Redis.TxPipeline()
	pipe.Incr(cctx, key)
	pipe.Expire(cctx, key, r.Interval)
	_, _ = pipe.Exec(cctx)

	return true, ""
}
