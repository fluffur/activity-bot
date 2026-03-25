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

func (r *StatsRepository) ChatMemberMessageStatsByChat(ctx context.Context, chatID int64, from, to *time.Time) ([]model.ChatMemberMessageCount, error) {

	var fromPg, toPg pgtype.Timestamptz

	if from != nil {
		fromPg = pgtype.Timestamptz{Time: *from, Valid: true}
	}
	if to != nil {
		toPg = pgtype.Timestamptz{Time: *to, Valid: true}
	}

	rows, err := r.queries.ChatMemberMessageStatsByChat(ctx, db.ChatMemberMessageStatsByChatParams{
		ChatID:   chatID,
		FromDate: fromPg,
		ToDate:   toPg,
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.ChatMemberMessageCount, len(rows))
	for i, row := range rows {
		result[i] = mapMessageReportRow(row)
	}

	return result, nil
}

func (r *StatsRepository) ChatMemberMessageStatsByUser(ctx context.Context, chatID int64, userID int64) (model.ChatMemberStats, error) {
	row, err := r.queries.ChatMemberMessageStatsByUser(ctx, db.ChatMemberMessageStatsByUserParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return model.ChatMemberStats{}, err
	}

	return mapMessageReportOneRow(row), nil
}

func (r *StatsRepository) UserMessageActivityDaily(ctx context.Context, chatID int64, userID int64) ([]model.MessageActivity, error) {
	activities, err := r.queries.UserMessageActivityDaily(ctx, db.UserMessageActivityDailyParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.MessageActivity, len(activities))
	for i, activity := range activities {
		result[i] = mapMessageActivity(db.ChatMessageActivityDailyRow(activity))
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
			Member:       mapInactiveChatMembersRow(member),
			LastActivity: member.LastMessageAt.Time,
		}
	}
	return result, nil
}
func (r *StatsRepository) ChatMessageActivityDaily(ctx context.Context, chatID int64, from, to *time.Time) ([]model.MessageActivity, error) {
	var fromPg, toPg pgtype.Timestamptz
	if from != nil {
		fromPg = pgtype.Timestamptz{Time: *from, Valid: true}
	}
	if to != nil {
		toPg = pgtype.Timestamptz{Time: *to, Valid: true}
	}

	rows, err := r.queries.ChatMessageActivityDaily(ctx, db.ChatMessageActivityDailyParams{
		ChatID:   chatID,
		FromDate: fromPg,
		ToDate:   toPg,
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.MessageActivity, len(rows))
	for i, row := range rows {
		result[i] = mapMessageActivity(row)
	}

	return result, nil
}
