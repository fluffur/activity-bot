package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"activity-bot/internal/stats"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type StatsRepository struct {
	queries *db.Queries
}

func NewStatsRepository(queries *db.Queries) stats.Repository {
	return &StatsRepository{queries}
}

func (r *StatsRepository) GetWeeklyReport(ctx context.Context, chatID int64) ([]model.WeeklyMessageReportMember, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	daysSinceMonday := (weekday + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	rows, err := r.queries.WeeklyMessageReport(ctx, db.WeeklyMessageReportParams{
		ChatID: chatID,
		CreatedAt: pgtype.Timestamptz{
			Time:  monday,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.WeeklyMessageReportMember, len(rows))
	for i, row := range rows {
		result[i] = mapWeeklyReportRow(row)
	}

	return result, nil
}

func mapWeeklyReportRow(row db.WeeklyMessageReportRow) model.WeeklyMessageReportMember {
	fullName := row.FirstName.String
	if row.LastName.Valid {
		fullName += " " + row.LastName.String
	}
	var username *string
	if row.Username.Valid {
		username = &row.Username.String
	}

	return model.WeeklyMessageReportMember{
		UserID:        row.UserID,
		FullName:      fullName,
		MessagesCount: int32(row.MessagesCount),
		WeeklyNorm:    row.WeeklyNorm,
		NormDone:      row.NormDone,
		Username:      username,
	}
}
