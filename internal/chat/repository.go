package chat

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	Ensure(ctx context.Context, c model.Chat) (model.Chat, error)
	SetWarnNorm(ctx context.Context, chatID int64, norm int32) error
	SetBanNorm(ctx context.Context, chatID int64, norm int32) error
	SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error
	GetNorm(ctx context.Context, chatID int64, fallbackNorm int32) (int, error)
	GetNewbieThreshold(ctx context.Context, chatID int64) (int, error)
	GetChat(ctx context.Context, chatID int64) (model.Chat, error)
	SetChatPrompt(ctx context.Context, chatID int64, prompt string) error
	SetMaxLadder(ctx context.Context, chatID int64, maxLadder int32) error
	SetWelcomeCallMessage(ctx context.Context, chatID int64, message string) error
	UpdateCallOnJoin(ctx context.Context, chatID int64, isEnabled bool) error
	SetWeekStartDay(ctx context.Context, chatID int64, day int) error
}
