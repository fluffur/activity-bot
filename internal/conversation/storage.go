package conversation

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Storage interface {
	Get(ctx context.Context, chatID, userID int64) (string, error)
	Set(ctx context.Context, chatID, userID int64, state string, ttl time.Duration) error
	Delete(ctx context.Context, chatID, userID int64) error
}

type RedisStorage struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisStorage(rdb *redis.Client, prefix string) *RedisStorage {
	return &RedisStorage{
		rdb:    rdb,
		prefix: prefix,
	}
}

func (r *RedisStorage) key(chatID, userID int64) string {
	return fmt.Sprintf("%s:%d:%d", r.prefix, chatID, userID)
}

func (r *RedisStorage) Get(ctx context.Context, chatID, userID int64) (string, error) {
	val, err := r.rdb.Get(ctx, r.key(chatID, userID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisStorage) Set(ctx context.Context, chatID, userID int64, state string, ttl time.Duration) error {
	return r.rdb.Set(ctx, r.key(chatID, userID), state, ttl).Err()
}

func (r *RedisStorage) Delete(ctx context.Context, chatID, userID int64) error {
	return r.rdb.Del(ctx, r.key(chatID, userID)).Err()
}
