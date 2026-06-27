package stats

import (
	"context"
	"time"
)

type ProfileStatsRange struct {
	DayStart          time.Time
	DayRollingStart   time.Time
	WeekStart         time.Time
	WeekRollingStart  time.Time
	MonthStart        time.Time
	MonthRollingStart time.Time
}

type Repository interface {
	ChatStats(ctx context.Context, chatID int64, fromDate time.Time, toDate time.Time) ([]ChatStats, error)
	ProfileStats(ctx context.Context, chatID, userID int64, statsRange ProfileStatsRange) (ProfileStats, error)
}
