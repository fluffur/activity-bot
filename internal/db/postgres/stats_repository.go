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
	return model.MemberStats{
		User: model.User{
			ID:        row.UserID,
			FirstName: row.FirstName.String,
			LastName:  row.LastName.String,
			Username:  row.Username.String,
		},
		DayCount:          int(row.DayCount),
		DayRollingCount:   int(row.DayRollingCount),
		WeekCount:         int(row.WeekCount),
		WeekRollingCount:  int(row.WeekRollingCount),
		MonthCount:        int(row.MonthCount),
		MonthRollingCount: int(row.MonthRollingCount),
		AllTime:           int(row.AllTimeCount),

		RestUntil: row.RestUntil.Time,

		NormBan:         int(row.NormBan.Int32),
		NormWarn:        int(row.NormWarn),
		JoinedAt:        row.JoinedAt.Time,
		NewbieThreshold: int(row.NewbieThresholdDays),
		Status:          row.Status,
		CustomTitle:     row.CustomTitle.String,
		LeftAt:          row.LeftAt.Time,
	}
}

func mapWeeklyReportRow(row db.MessageReportRow) model.MessageReportMember {
	return model.MessageReportMember{
		User: model.User{
			ID:        row.UserID,
			FirstName: row.FirstName.String,
			Username:  row.Username.String,
		},
		MessagesCount:       int(row.MessagesCount),
		NormWarn:            int(row.NormWarn),
		NormBan:             int(row.NormBan.Int32),
		JoinedAt:            row.JoinedAt.Time,
		NewbieThresholdDays: int(row.NewbieThresholdDays),
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
		result[i] = model.InactiveMember{
			Member: model.ChatMember{
				User: model.User{
					ID:        member.ID,
					FirstName: member.FirstName.String,
					LastName:  member.LastName.String,
					Username:  member.Username.String,
				},
				RestUntil:   member.RestUntil.Time,
				CustomTitle: member.CustomTitle.String,
				Status:      member.Status,
			},
			LastActivity: member.LastMessageAt.Time,
		}
	}
	return result, nil
}
func (r *StatsRepository) GetMessageActivityByDayAll(ctx context.Context, chatID int64, from, to *time.Time) ([]model.MessageActivity, error) {
	var fromPg, toPg pgtype.Timestamptz
	if from != nil {
		fromPg = pgtype.Timestamptz{Time: *from, Valid: true}
	}
	if to != nil {
		toPg = pgtype.Timestamptz{Time: *to, Valid: true}
	}

	rows, err := r.queries.MessageActivityByDayAll(ctx, db.MessageActivityByDayAllParams{
		ChatID:   chatID,
		FromDate: fromPg,
		ToDate:   toPg,
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.MessageActivity, len(rows))
	for i, row := range rows {
		result[i] = mapMessageActivity(db.MessageActivityByDayRow(row))
	}

	return result, nil
}
