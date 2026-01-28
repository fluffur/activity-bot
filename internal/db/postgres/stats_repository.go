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

func (r *StatsRepository) GetReport(
	ctx context.Context,
	chatID int64,
	from, to *time.Time,
) ([]model.MessageReportMember, error) {

	var fromPg, toPg pgtype.Timestamptz

	if from != nil {
		fromPg = pgtype.Timestamptz{Time: *from, Valid: true}
	}
	if to != nil {
		toPg = pgtype.Timestamptz{Time: *to, Valid: true}
	}

	rows, err := r.queries.MessageReport(ctx, db.MessageReportParams{
		ChatID:   chatID,
		FromDate: fromPg,
		ToDate:   toPg,
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.MessageReportMember, len(rows))
	for i, row := range rows {
		result[i] = mapWeeklyReportRow(row)
	}

	return result, nil
}

func mapWeeklyReportRow(row db.MessageReportRow) model.MessageReportMember {
	var username *string
	if row.Username.Valid {
		username = &row.Username.String
	}

	return model.MessageReportMember{
		User: model.User{
			ID:        row.UserID,
			FirstName: row.FirstName.String,
			Username:  username,
		},
		MessagesCount:       int32(row.MessagesCount),
		WeeklyNorm:          row.WeeklyNorm,
		NormDone:            row.NormDone,
		JoinedAt:            row.JoinedAt.Time,
		NewbieThresholdDays: row.NewbieThresholdDays,
		CustomTitle:         row.CustomTitle.String,
		Role:                row.Role,
	}
}
