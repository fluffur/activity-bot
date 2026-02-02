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

func (r *StatsRepository) GetReport(ctx context.Context, chatID int64, from, to *time.Time) ([]model.MessageReportMember, error) {

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

func (r *StatsRepository) GetReportOne(ctx context.Context, chatID int64, userID int64) (model.MemberStats, error) {
	row, err := r.queries.MessageReportOne(ctx, db.MessageReportOneParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return model.MemberStats{}, err
	}

	return mapMemberStats(row), nil
}

func mapMemberStats(row db.MessageReportOneRow) model.MemberStats {
	var username *string
	if row.Username.Valid {
		username = &row.Username.String
	}

	var customTitle *string
	if row.CustomTitle.Valid {
		customTitle = &row.CustomTitle.String
	}
	var restUntil *time.Time
	if row.RestUntil.Valid {
		restUntil = &row.RestUntil.Time
	}
	return model.MemberStats{
		User: model.User{
			ID:        row.UserID,
			FirstName: row.FirstName.String,
			Username:  username,
		},
		DayCount:          int32(row.DayCount),
		WeekCount:         int32(row.WeekCount),
		WeekRollingCount:  int32(row.WeekRollingCount),
		MonthCount:        int32(row.MonthCount),
		MonthRollingCount: int32(row.MonthRollingCount),
		AllTime:           int32(row.AllTimeCount),

		RestUntil: restUntil,

		WeeklyNorm:      row.WeeklyNorm,
		JoinedAt:        row.JoinedAt.Time,
		NewbieThreshold: row.NewbieThresholdDays,
		Status:          row.Status,
		CustomTitle:     customTitle,
	}
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
		Status:              row.Status,
	}
}
