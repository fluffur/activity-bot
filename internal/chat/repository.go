package chat

import "context"

type Repository interface {
	Create(ctx context.Context, chat Chat) error
	GetByID(ctx context.Context, chatID int64) (Chat, error)
}
