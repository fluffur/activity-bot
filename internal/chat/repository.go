package chat

import "context"

type Repository interface {
	Create(ctx context.Context, chat Chat) error
	Get(ctx context.Context, chatID int64) (Chat, error)
}
