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
		DayRollingCount:   int32(row.DayRollingCount),
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

func mapMessageActivity(row db.MessageActivityByDayRow) model.MessageActivity {
	return model.MessageActivity{
		Count: row.MessagesCount,
		Date:  row.Day.Time,
	}
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

func (r *StatsRepository) GetMessageActivityByDay(ctx context.Context, chatID int64, userID int64) ([]model.MessageActivity, error) {
	activities, err := r.queries.MessageActivityByDay(ctx, db.MessageActivityByDayParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.MessageActivity, len(activities))
	for i, activity := range activities {
		result[i] = mapMessageActivity(activity)
	}

	return result, nil
}

func (r *StatsRepository) GetInactiveMembers(ctx context.Context, chatID int64) ([]model.InactiveMember, error) {
	members, err := r.queries.InactiveChatMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make([]model.InactiveMember, len(members))
	for i, member := range members {
		var lastMessageAt *time.Time
		if member.LastMessageAt.Valid {
			lastMessageAt = &member.LastMessageAt.Time
		}
		var username *string
		if member.Username.Valid {
			username = &member.Username.String
		}
		var restUntil *time.Time
		if member.RestUntil.Valid {
			restUntil = &member.RestUntil.Time
		}

		result[i] = model.InactiveMember{
			Member: model.ChatMember{
				User: model.User{
					ID:        member.ID,
					FirstName: member.FirstName.String,
					LastName:  member.LastName.String,
					Username:  username,
				},
				RestUntil:   restUntil,
				CustomTitle: member.CustomTitle.String,
				Status:      member.Status,
			},
			LastActivity: lastMessageAt,
		}
	}
	return result, nil
}
