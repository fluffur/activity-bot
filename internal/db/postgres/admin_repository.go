package postgres

import (
	"activity-bot/internal/admin"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"context"
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
