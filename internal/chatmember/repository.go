package chatmember

import (
	"activity-bot/internal/permission"
	"context"
	"time"
)

type OptionalBool struct {
	Bool  bool
	Valid bool
}

type Filter struct {
	ChatID   int64
	IsBot    OptionalBool
	Left     OptionalBool
	Excluded OptionalBool
}

type Repository interface {
	Create(ctx context.Context, chatMember ChatMember) error
	Get(ctx context.Context, chatID int64, userID int64) (ChatMember, error)
	GetByUsername(ctx context.Context, chatID int64, username string) (ChatMember, error)
	SetTag(ctx context.Context, chatID int64, userID int64, tag string) error
	MarkLeft(ctx context.Context, chatID int64, userID int64, leftAt time.Time) error
	Restore(ctx context.Context, chatID int64, userID int64) error
	MarkAllLeftExcept(ctx context.Context, chatID int64, userIDs []int64, leftAt time.Time) error
	UpsertChatMembers(ctx context.Context, chatID int64, chatMembers []ChatMember) error
	List(ctx context.Context, filter Filter) ([]ChatMember, error)
	ListAdmins(ctx context.Context, chatID int64, minStatus permission.Status) ([]ChatMember, error)
	SetExcludeFromSummon(ctx context.Context, chatID, userID int64, excluded bool) error
	SetEmoji(c context.Context, chatID, userID int64, emojisString string) error
	SetDescription(ctx context.Context, chatID, userID int64, description string) error
}
