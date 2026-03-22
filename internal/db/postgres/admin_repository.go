package postgres

import (
	"activity-bot/internal/admin"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type AdminRepository struct {
	queries *db.Queries
}

func NewAdminRepository(queries *db.Queries) admin.Repository {
	return &AdminRepository{queries}
}

func (r *AdminRepository) SetStatus(ctx context.Context, chatID int64, userID int64, status int16) error {
	return r.queries.SetChatMemberStatus(ctx, db.SetChatMemberStatusParams{
		ChatID: chatID,
		UserID: userID,
		Status: status,
	})
}

func (r *AdminRepository) GetAdmins(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	rows, err := r.queries.GetChatAdmins(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return mapMembers(rows, func(row db.GetChatAdminsRow) model.ChatMember {
		return mapChatMemberFull(row.ChatMember, row.User)
	}), nil
}

func (r *AdminRepository) CreateModerationAction(ctx context.Context, actionType string, chatID, userID, modID int64, reason string, until *time.Time) error {
	var expiresAt pgtype.Timestamptz
	if until != nil {
		expiresAt = pgtype.Timestamptz{
			Time:  *until,
			Valid: true,
		}
	}

	return r.queries.CreateModerationAction(ctx, db.CreateModerationActionParams{
		Type:        db.ModerationType(actionType),
		ChatID:      chatID,
		UserID:      userID,
		ModeratorID: modID,
		Reason:      pgtype.Text{String: reason, Valid: reason != ""},
		ExpiresAt:   expiresAt,
	})
}

func (r *AdminRepository) GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error) {
	return r.queries.GetActiveWarnsCount(ctx, db.GetActiveWarnsCountParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) GetActiveWarns(ctx context.Context, chatID, userID int64) ([]model.Warn, error) {
	warns, err := r.queries.GetActiveWarns(ctx, db.GetActiveWarnsParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	results := make([]model.Warn, len(warns))
	for i, warn := range warns {
		results[i] = model.Warn{
			ID:         warn.ID,
			Moderator:  mapChatMemberFull(warn.ChatMember, warn.User),
			ChatMember: mapChatMemberFull(warn.ChatMember_2, warn.User_2),
			Reason:     warn.Reason.String,
			CreatedAt:  warn.CreatedAt.Time,
			ExpiresAt:  warn.ExpiresAt.Time,
		}
	}
	return results, nil
}

func (r *AdminRepository) ClearWarns(ctx context.Context, chatID, userID int64) error {
	return r.queries.ClearWarns(ctx, db.ClearWarnsParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) GetChatMaxWarns(ctx context.Context, chatID int64) (int, error) {
	maxWarns, err := r.queries.GetChatMaxWarns(ctx, chatID)
	return int(maxWarns), err
}

func (r *AdminRepository) UpdateChatMaxWarns(ctx context.Context, chatID int64, maxWarns int) error {
	return r.queries.UpdateChatMaxWarns(ctx, db.UpdateChatMaxWarnsParams{
		MaxWarns: int32(maxWarns),
		ID:       chatID,
	})
}

func (r *AdminRepository) RemoveModerationActions(ctx context.Context, chatID, userID int64) error {
	return r.queries.DeleteModerationActionsForUser(ctx, db.DeleteModerationActionsForUserParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) RemoveLatestWarn(ctx context.Context, chatID, userID int64) error {
	return r.queries.RemoveLatestWarn(ctx, db.RemoveLatestWarnParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) EnsureDeveloperUser(ctx context.Context, userID int64) error {
	return r.queries.EnsureDeveloperUser(ctx, userID)
}

func (r *AdminRepository) GetActiveWarnsByChat(ctx context.Context, chatID int64) ([]model.Warn, error) {
	warns, err := r.queries.GetActiveWarnsByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	results := make([]model.Warn, len(warns))
	for i, warn := range warns {
		results[i] = model.Warn{
			ID:         warn.ID,
			Moderator:  mapChatMemberFull(warn.ChatMember, warn.User),
			ChatMember: mapChatMemberFull(warn.ChatMember_2, warn.User_2),
			Reason:     warn.Reason.String,
			CreatedAt:  warn.CreatedAt.Time,
			ExpiresAt:  warn.ExpiresAt.Time,
		}
	}
	return results, nil
}
