package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const applicationTTL = 24 * time.Hour

type Repository struct {
	redis *redis.Client
}

func NewRepository(redis *redis.Client) *Repository {
	return &Repository{
		redis: redis,
	}
}

func (r *Repository) Save(ctx context.Context, app Application) error {
	data, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("marshal application: %w", err)
	}

	if err := r.redis.Set(
		ctx,
		keyApplication(app.ChatID, app.UserID),
		data,
		applicationTTL,
	).Err(); err != nil {
		return fmt.Errorf("save application: %w", err)
	}

	return nil
}

func (r *Repository) Get(ctx context.Context, chatID, userID int64) (*Application, error) {
	data, err := r.redis.Get(ctx, keyApplication(chatID, userID)).Bytes()

	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}

	var app Application

	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("unmarshal application: %w", err)
	}

	return &app, nil
}

func (r *Repository) Delete(ctx context.Context, chatID, userID int64) error {
	if err := r.redis.Del(ctx, keyApplication(chatID, userID)).Err(); err != nil {
		return fmt.Errorf("delete application: %w", err)
	}

	return nil
}

func keyApplication(chatID, userID int64) string {
	return fmt.Sprintf(
		"application:%d:%d",
		chatID,
		userID,
	)
}
