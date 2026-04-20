package stats

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	ChatMemberMessageStatsByChat(ctx context.Context, chatID int64, from, to *time.Time) ([]model.ChatMemberMessageCount, error)
	ChatMemberMessageStatsByUser(ctx context.Context, chatID int64, userID int64, from time.Time) (model.ChatMemberStats, error)
	UserMessageActivityDaily(ctx context.Context, chatID int64, userID int64) ([]model.MessageActivity, error)
	GetInactiveMembers(ctx context.Context, chatID int64) ([]model.InactiveMember, error)
	ChatMessageActivityDaily(ctx context.Context, chatID int64, from *time.Time, to *time.Time) ([]model.MessageActivity, error)
}
