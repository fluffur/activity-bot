package redis

import (
	"activity-bot/internal/crocodile"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const gameTTL = 6 * time.Hour

type CrocodileRepository struct {
	rdb *redis.Client
}

func NewCrocodileRepository(rdb *redis.Client) *CrocodileRepository {
	return &CrocodileRepository{
		rdb: rdb,
	}
}

func (r *CrocodileRepository) key(chatID int64) string {
	return fmt.Sprintf("crocodile:game:%d", chatID)
}

func (r *CrocodileRepository) Create(ctx context.Context, game *crocodile.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, r.key(game.ChatID), data, gameTTL).Err()
}

func (r *CrocodileRepository) Get(ctx context.Context, chatID int64) (*crocodile.Game, error) {
	data, err := r.rdb.Get(ctx, r.key(chatID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var game crocodile.Game
	if err := json.Unmarshal(data, &game); err != nil {
		return nil, err
	}

	return &game, nil
}

func (r *CrocodileRepository) Update(ctx context.Context, game *crocodile.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}

	ttl, err := r.rdb.TTL(ctx, r.key(game.ChatID)).Result()
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = gameTTL
	}

	return r.rdb.Set(ctx, r.key(game.ChatID), data, ttl).Err()
}

func (r *CrocodileRepository) Delete(ctx context.Context, chatID int64) error {
	return r.rdb.Del(ctx, r.key(chatID)).Err()
}

func (r *CrocodileRepository) Exists(ctx context.Context, chatID int64) (bool, error) {
	n, err := r.rdb.Exists(ctx, r.key(chatID)).Result()
	return n == 1, err
}
