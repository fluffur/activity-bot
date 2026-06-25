package stats

import (
	"context"
	"time"
)

type Repository interface {
	ChatStats(ctx context.Context, chatID int64, fromDate time.Time, toDate time.Time) ([]ChatStats, error)
}
