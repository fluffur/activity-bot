package admin

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	Add(ctx context.Context, chatID int64, userID int64) error
	Remove(ctx context.Context, chatID int64, userID int64) error
	GetFromChat(ctx context.Context, chatID int64) ([]model.User, error)
	IsAdmin(ctx context.Context, chatID int64, userID int64) (bool, error)
	IsCreator(ctx context.Context, chatID int64, userID int64) (bool, error)
	GetRole(ctx context.Context, chatID int64, userID int64) (string, error)
	GetChatsWhereUserIsAdmin(ctx context.Context, userID int64) ([]model.Chat, error)
	GetAllChats(ctx context.Context) ([]model.Chat, error)

	EnsureDeveloperUser(ctx context.Context, userID int64) error
	GetDeveloperRole(ctx context.Context, chatID, userID int64) (string, error)
	SetDeveloperRole(ctx context.Context, chatID, userID int64, role string) error
	RemoveDeveloperRole(ctx context.Context, chatID, userID int64) error
	IsDeveloper(ctx context.Context, chatID, userID int64) (bool, error)
	GetAllDevelopers(ctx context.Context, chatID int64) ([]model.User, []string, error)
	CreateModerationAction(ctx context.Context, actionType string, chatID, userID, modID int64, reason string, until *time.Time) error
	GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error)
	ClearWarns(ctx context.Context, chatID, userID int64) error
	GetChatMaxWarns(ctx context.Context, chatID int64) (int, error)
	UpdateChatMaxWarns(ctx context.Context, chatID int64, maxWarns int) error
	RemoveModerationActions(ctx context.Context, chatID, userID int64) error
	RemoveLatestWarn(ctx context.Context, chatID, userID int64) error
	GetActiveWarns(ctx context.Context, chatID, userID int64) ([]model.Warn, error)
	GetActiveWarnsByChat(ctx context.Context, chatID int64) ([]model.Warn, error)
	GetChatsWithoutTitle(ctx context.Context) ([]model.Chat, error)
}
