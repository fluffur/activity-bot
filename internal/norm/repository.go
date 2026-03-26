package norm

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, chatID int64, name string, value int32) error
	Get(ctx context.Context, chatID int64, name string) (ChatNorm, error)
	ListAll(ctx context.Context, chatID int64) ([]ChatNorm, error)
}
