package chat

import "context"

type Repository interface {
	Create(ctx context.Context, chat Chat) error
	Get(ctx context.Context, chatID int64) (Chat, error)
	SetMentionTypes(ctx context.Context, chatID int64, mentionTypes MentionTypes) error
	SetSkipSummonConfirmation(ctx context.Context, chatID int64, confirmation bool) error
	GetUserManagedChats(ctx context.Context, userID int64, search string) ([]Chat, error)
	GetAllChats(ctx context.Context, search string) ([]Chat, error)
}
