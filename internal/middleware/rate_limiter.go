package middleware

import (
	"activity-bot/internal/command"
	"activity-bot/internal/options"
	"errors"
	"fmt"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Redis    *redis.Client
	Limit    int
	Interval time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, interval time.Duration) command.Middleware {
	return &RateLimiter{
		Redis:    rdb,
		Limit:    limit,
		Interval: interval,
	}
}

func (r *RateLimiter) CheckUpdate(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	key := fmt.Sprintf("rate:%d:%s", c.ID, ctx.Command.Name())
	count, err := r.Redis.Get(ctx.StdContext(), key).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil
	}

	if count >= r.Limit {
		ttl, err := r.Redis.TTL(ctx.StdContext(), key).Result()
		if err != nil || ttl < 0 {
			ttl = r.Interval
		}
		seconds := int(ttl.Seconds())

		if err := ctx.ReplyOnly(u, options.WithText(fmt.Sprintf("⚠️ Слишком много запросов! Попробуйте через %d секунд.", seconds))); err != nil {
			return err
		}

		return command.ErrStop
	}

	pipe := r.Redis.TxPipeline()
	pipe.Incr(ctx.StdContext(), key)
	pipe.Expire(ctx.StdContext(), key, r.Interval)
	_, _ = pipe.Exec(ctx.StdContext())

	return nil
}
