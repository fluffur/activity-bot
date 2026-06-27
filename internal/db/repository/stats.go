package repository

import (
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/stats"
	"context"
	"time"
)

type StatsRepository struct {
	queries *db.Queries
}

func NewStatsRepository(q *db.Queries) *StatsRepository {
	return &StatsRepository{queries: q}
}

func (r *StatsRepository) ChatStats(ctx context.Context, chatID int64, fromDate, toDate time.Time) ([]stats.ChatStats, error) {
	chatStats, err := r.queries.ChatStats(ctx, db.ChatStatsParams{
		FromDate: timestamptz(fromDate),
		ToDate:   timestamptz(toDate),
		ChatID:   chatID,
	})
	if err != nil {
		return nil, err
	}
	return mapList(chatStats, func(t db.ChatStatsRow) stats.ChatStats {
		return mapChatStats(t.MessagesCount, t.ChatMember, t.User)
	}), nil
}

func (r *StatsRepository) ProfileStats(
	ctx context.Context, chatID, userID int64, statsRange stats.ProfileStatsRange,
) (stats.ProfileStats, error) {
	s, err := r.queries.UserStats(ctx, db.UserStatsParams{
		DayStart:          timestamptz(statsRange.DayStart),
		DayRollingStart:   timestamptz(statsRange.DayRollingStart),
		WeekStart:         timestamptz(statsRange.WeekStart),
		WeekRollingStart:  timestamptz(statsRange.WeekRollingStart),
		MonthStart:        timestamptz(statsRange.MonthStart),
		MonthRollingStart: timestamptz(statsRange.MonthRollingStart),
		ChatID:            chatID,
		UserID:            userID,
	})
	if err != nil {
		return stats.ProfileStats{}, err
	}

	return mapProfileStats(s), nil
}
