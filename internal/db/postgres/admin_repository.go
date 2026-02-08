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

func (r *AdminRepository) Add(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.AddChatAdmin(ctx, db.AddChatAdminParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) Remove(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.RemoveChatAdmin(ctx, db.RemoveChatAdminParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) GetFromChat(ctx context.Context, chatID int64) ([]model.User, error) {
	rows, err := r.queries.GetChatAdmins(ctx, chatID)
	if err != nil {
		return nil, err
	}

	admins := make([]model.User, len(rows))
	for i, row := range rows {
		var username *string
		if row.Username.Valid {
			username = &row.Username.String
		}
		admins[i] = model.User{
			ID:        row.ID,
			FirstName: row.FirstName.String,
			Username:  username,
		}
	}
	return admins, nil
}

func (r *AdminRepository) IsAdmin(ctx context.Context, chatID int64, userID int64) (bool, error) {
	return r.queries.IsChatAdmin(ctx, db.IsChatAdminParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) IsCreator(ctx context.Context, chatID int64, userID int64) (bool, error) {
	return r.queries.IsChatCreator(ctx, db.IsChatCreatorParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) GetRole(ctx context.Context, chatID int64, userID int64) (string, error) {
	return r.queries.GetChatMemberStatus(ctx, db.GetChatMemberStatusParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *AdminRepository) CreateModerationAction(ctx context.Context, actionType string, chatID, userID, modID int64, reason string, until *time.Time) error {
	var untilDate pgtype.Timestamp
	if until != nil {
		untilDate = pgtype.Timestamp{
			Time:  *until,
			Valid: true,
		}
	}

	return r.queries.CreateModerationAction(ctx, db.CreateModerationActionParams{
		Type:      db.ModerationType(actionType),
		ChatID:    chatID,
		UserID:    userID,
		ModID:     modID,
		Reason:    pgtype.Text{String: reason, Valid: reason != ""},
		UntilDate: untilDate,
	})
}

func (r *AdminRepository) GetWarnsCount(ctx context.Context, chatID, userID int64) (int64, error) {
	return r.queries.GetActiveWarnsCount(ctx, db.GetActiveWarnsCountParams{
		ChatID: chatID,
		UserID: userID,
	})
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
