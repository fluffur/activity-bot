package member

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	GetCustomTitle(ctx context.Context, chatID int64, userID int64) (string, error)
	UpdateCustomTitle(ctx context.Context, chatID int64, userID int64, title string) error
	FindByChatID(ctx context.Context, chatID int64) ([]model.ChatMember, error)
	GetWithCustomTitles(ctx context.Context, chatID int64) ([]model.ChatMember, error)
	UpsertChatMembers(ctx context.Context, chatID int64, users []model.ChatMemberUpdate) error
	Get(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error)
	Remove(ctx context.Context, chatID int64, userID int64) error
	EnsureExists(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error)
}
