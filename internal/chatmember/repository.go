package chatmember

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, chatMember ChatMember) error
	Get(ctx context.Context, chatID int64, userID int64) (ChatMember, error)
	SetTag(ctx context.Context, chatID int64, userID int64, tag string) error
	MarkLeft(ctx context.Context, chatID int64, userID int64, leftAt time.Time) error
	MarkAllLeftExcept(ctx context.Context, chatID int64, userIDs []int64, leftAt time.Time) error
	UpsertChatMembers(ctx context.Context, chatID int64, chatMembers []ChatMember) error
}
