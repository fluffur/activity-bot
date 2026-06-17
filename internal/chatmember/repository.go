package chatmember

import "context"

type Repository interface {
	Create(ctx context.Context, chatMember ChatMember) error
	Get(ctx context.Context, chatID int64, userID int64) (ChatMember, error)
}
