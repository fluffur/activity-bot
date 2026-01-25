package admin

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	Add(ctx context.Context, chatID int64, userID int64) error
	Remove(ctx context.Context, chatID int64, userID int64) error
	GetFromChat(ctx context.Context, chatID int64) ([]model.User, error)
	IsAdmin(ctx context.Context, chatID int64, userID int64) (bool, error)
	IsCreator(ctx context.Context, chatID int64, userID int64) (bool, error)
	GetRole(ctx context.Context, chatID int64, userID int64) (string, error)
}
