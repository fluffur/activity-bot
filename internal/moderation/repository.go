package moderation

import (
	"activity-bot/internal/chatmember"
	"context"
	"time"
)

type Warn struct {
	ID        int64
	Target    chatmember.ChatMember
	Moderator chatmember.ChatMember
	Reason    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Repository interface {
	SetStatus(ctx context.Context, chatID int64, userID int64, status int16) error
	CreateModerationAction(ctx context.Context, actionType string, chatID, userID, modID int64, reason string, until time.Time) error
	RemoveModerationActions(ctx context.Context, chatID, userID int64) error
	RemoveLatestWarn(ctx context.Context, chatID, userID int64) error
	GetActiveWarns(ctx context.Context, chatID, userID int64) ([]Warn, error)
	GetActiveWarnsByChat(ctx context.Context, chatID int64) ([]Warn, error)
	GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error)
	ClearWarns(ctx context.Context, chatID, userID int64) error
	GetChatMaxWarns(ctx context.Context, chatID int64) (int, error)
	UpdateChatMaxWarns(ctx context.Context, chatID int64, maxWarns int) error
}
