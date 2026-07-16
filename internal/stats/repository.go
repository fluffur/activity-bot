package stats

import (
	"activity-bot/internal/chatmember"
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

type InactiveMember struct {
	ChatMember   chatmember.ChatMember
	LastActivity time.Time
}

type Repository interface {
	ChatStats(ctx context.Context, chatID int64, fromDate time.Time, toDate time.Time) ([]ChatStats, error)
	ProfileStats(ctx context.Context, chatID, userID int64, statsRange ProfileStatsRange) (ProfileStats, error)
	ListInactiveMembers(ctx context.Context, chatID int64) ([]InactiveMember, error)
}
