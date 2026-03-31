package session

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	SetSession(ctx context.Context, userID int64, chatID int64) error
	GetSession(ctx context.Context, userID int64) (int64, error)
	GetChat(ctx context.Context, userID int64) (model.Chat, error)
	DeleteSession(ctx context.Context, userID int64) error
}
