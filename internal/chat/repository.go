package chat

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	Ensure(ctx context.Context, c model.Chat) (model.Chat, error)
	SetNorm(ctx context.Context, chatID int64, norm int32) error
	SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error
	GetNorm(ctx context.Context, chatID int64, fallbackNorm int32) (int, error)
	GetNewbieThreshold(ctx context.Context, chatID int64) (int, error)
	SetOnlyNewbies(ctx context.Context, chatID int64, users []*model.User) error
}
