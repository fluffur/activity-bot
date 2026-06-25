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

func (r *StatsRepository) ChatStats(ctx context.Context, chatID int64, fromDate time.Time, toDate time.Time) ([]stats.ChatStats, error) {
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
