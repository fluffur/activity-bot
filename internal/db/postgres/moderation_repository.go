package postgres

import (
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/moderation"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type ModerationRepository struct {
	queries *db.Queries
}

func NewModerationRepository(queries *db.Queries) *ModerationRepository {
	return &ModerationRepository{queries}
}

func (r *ModerationRepository) SetStatus(ctx context.Context, chatID int64, userID int64, status int16) error {
	return r.queries.SetChatMemberStatus(ctx, db.SetChatMemberStatusParams{
		ChatID: chatID,
		UserID: userID,
		Status: status,
	})
}

func (r *ModerationRepository) GetAdmins(ctx context.Context, chatID int64) ([]chatmember.ChatMember, error) {
	rows, err := r.queries.GetChatAdmins(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return mapList(rows, func(row db.GetChatAdminsRow) chatmember.ChatMember {
		return mapChatMemberFull(row.ChatMember, db.Chat{}, row.User)
	}), nil
}

func (r *ModerationRepository) CreateModerationAction(ctx context.Context, actionType string, chatID, userID, modID int64, reason string, until time.Time) error {
	return r.queries.CreateModerationAction(ctx, db.CreateModerationActionParams{
		Type:        db.ModerationType(actionType),
		ChatID:      chatID,
		UserID:      userID,
		ModeratorID: modID,
		Reason:      pgtype.Text{String: reason, Valid: reason != ""},
		ExpiresAt: pgtype.Timestamptz{
			Time:  until,
			Valid: !until.IsZero(),
		},
	})
}

func (r *ModerationRepository) GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error) {
	return r.queries.GetActiveWarnsCount(ctx, db.GetActiveWarnsCountParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ModerationRepository) GetActiveWarns(ctx context.Context, chatID, userID int64) ([]moderation.Warn, error) {
	warns, err := r.queries.GetActiveWarns(ctx, db.GetActiveWarnsParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	results := make([]moderation.Warn, len(warns))
	for i, warn := range warns {
		results[i] = moderation.Warn{
			ID:         warn.ID,
			Moderator:  mapChatMemberFull(warn.ChatMember, db.Chat{}, warn.User),
			ChatMember: mapChatMemberFull(warn.ChatMember_2, db.Chat{}, warn.User_2),
			Reason:     warn.Reason.String,
			CreatedAt:  warn.CreatedAt.Time,
			ExpiresAt:  warn.ExpiresAt.Time,
		}
	}
	return results, nil
}

func (r *ModerationRepository) ClearWarns(ctx context.Context, chatID, userID int64) error {
	return r.queries.ClearWarns(ctx, db.ClearWarnsParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ModerationRepository) GetChatMaxWarns(ctx context.Context, chatID int64) (int, error) {
	maxWarns, err := r.queries.GetChatMaxWarns(ctx, chatID)
	return int(maxWarns), err
}

func (r *ModerationRepository) UpdateChatMaxWarns(ctx context.Context, chatID int64, maxWarns int) error {
	return r.queries.UpdateChatMaxWarns(ctx, db.UpdateChatMaxWarnsParams{
		MaxWarns: int32(maxWarns),
		ID:       chatID,
	})
}

func (r *ModerationRepository) RemoveModerationActions(ctx context.Context, chatID, userID int64) error {
	return r.queries.DeleteModerationActionsForUser(ctx, db.DeleteModerationActionsForUserParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ModerationRepository) RemoveLatestWarn(ctx context.Context, chatID, userID int64) error {
	return r.queries.RemoveLatestWarn(ctx, db.RemoveLatestWarnParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ModerationRepository) EnsureDeveloperUser(ctx context.Context, userID int64) error {
	return r.queries.EnsureDeveloperUser(ctx, userID)
}

func (r *ModerationRepository) GetActiveWarnsByChat(ctx context.Context, chatID int64) ([]moderation.Warn, error) {
	warns, err := r.queries.GetActiveWarnsByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	results := make([]moderation.Warn, len(warns))
	for i, warn := range warns {
		results[i] = moderation.Warn{
			ID:         warn.ID,
			Moderator:  mapChatMemberFull(warn.ChatMember, db.Chat{}, warn.User),
			ChatMember: mapChatMemberFull(warn.ChatMember_2, db.Chat{}, warn.User_2),
			Reason:     warn.Reason.String,
			CreatedAt:  warn.CreatedAt.Time,
			ExpiresAt:  warn.ExpiresAt.Time,
		}
	}
	return results, nil
}
