package chat

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, chat Chat) error
	Get(ctx context.Context, chatID int64) (Chat, error)
	ListWithoutTitle(ctx context.Context) ([]Chat, error)
	SetMentionTypes(ctx context.Context, chatID int64, mentionTypes MentionTypes) error
	SetSkipSummonConfirmation(ctx context.Context, chatID int64, confirmation bool) error
	GetUserManagedChats(ctx context.Context, userID int64, search string) ([]Chat, error)
	GetAllChats(ctx context.Context, search string) ([]Chat, error)
	SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error
	SetTitle(ctx context.Context, chatID int64, title string) error
	Remove(ctx context.Context, chatID int64) error
	SetChatPrompt(ctx context.Context, chatID int64, prompt string) error
	SetWelcomeCallMessage(ctx context.Context, chatID int64, message string) error
	SetCallOnJoin(ctx context.Context, chatID int64, isEnabled bool) error
	SetWeekStartDay(ctx context.Context, chatID int64, day int) error
	SetWeekStartTime(ctx context.Context, chatID int64, timeMicroseconds int64) error
	SetEmojisEnabled(ctx context.Context, chatID int64, enabled bool) error
	SetAllowPrefixless(ctx context.Context, chatID int64, allow bool) error
	SetCommandPrefix(ctx context.Context, chatID int64, prefix string) error
}
