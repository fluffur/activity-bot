package norm

import (
	"activity-bot/internal/chatmember"
	"context"
)

type Repository interface {
	Get(ctx context.Context, chatID int64, name string) (Norm, error)
	GetNormMembers(ctx context.Context, normID int64) ([]chatmember.ChatMember, error)
	Set(ctx context.Context, chatID int64, name string, value int32) (int64, error)
	List(ctx context.Context, chatID int64) ([]Norm, error)
	Delete(ctx context.Context, normID int64) error
	Assign(ctx context.Context, normID int64, userIDs []int64) error
	Unassign(ctx context.Context, normID int64, userIDs []int64) error
}
