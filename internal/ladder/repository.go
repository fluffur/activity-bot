package ladder

import (
	"context"
	"time"
)

type Repository interface {
	Inc(ctx context.Context, chatID, userID int64, ttl time.Duration) (int64, bool, error)
	Reset(ctx context.Context, chatID int64) error
}
