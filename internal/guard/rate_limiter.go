package guard

//
//type RateLimiter struct {
//	Redis          *redis.Client
//	Limit          int
//	Interval       time.Duration
//	sessionService interface {
//		GetActiveChat(ctx context.Context, userID int64) (int64, error)
//	}
//}
//
//func NewRateLimiter(rdb *redis.Client, limit int, interval time.Duration, sessionService interface {
//	GetActiveChat(ctx context.Context, userID int64) (int64, error)
//}) command.Guard {
//	return &RateLimiter{
//		Redis:          rdb,
//		Limit:          limit,
//		Interval:       interval,
//		sessionService: sessionService,
//	}
//}
//
//func (r *RateLimiter) Check(ctx *ext.Context, command string, stdCtx context.Context) (bool, string) {
//	chatID := ctx.EffectiveChat.Id
//	if ctx.EffectiveChat.Type == "private" && r.sessionService != nil {
//		targetID, err := r.sessionService.GetActiveChat(stdCtx, ctx.EffectiveSender.Id())
//		if err == nil && targetID != 0 {
//			chatID = targetID
//		}
//	}
//
//	key := fmt.Sprintf("rate:%d:%s", chatID, command)
//	count, err := r.Redis.Get(stdCtx, key).Int()
//	if err != nil && !errors.Is(err, redis.Nil) {
//		return false, ""
//	}
//
//	if count >= r.Limit {
//		ttl, err := r.Redis.TTL(stdCtx, key).Result()
//		if err != nil || ttl < 0 {
//			ttl = r.Interval
//		}
//		seconds := int(ttl.Seconds())
//
//		return false, fmt.Sprintf("⚠️ Слишком много запросов! Попробуйте через %d секунд.", seconds)
//	}
//
//	pipe := r.Redis.TxPipeline()
//	pipe.Incr(stdCtx, key)
//	pipe.Expire(stdCtx, key, r.Interval)
//	_, _ = pipe.Exec(stdCtx)
//
//	return true, ""
//}
