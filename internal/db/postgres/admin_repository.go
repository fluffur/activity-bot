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

func (r *AdminRepository) GetFromChat(ctx context.Context, chatID int64) ([]model.ChatAdmin, error) {
	rows, err := r.queries.GetChatAdmins(ctx, chatID)
	if err != nil {
		return nil, err
	}

	admins := make([]model.ChatAdmin, len(rows))
	for i, row := range rows {
		displayName := row.Username.String
		if displayName == "" {
			displayName = row.FirstName.String
			if row.LastName.Valid {
				displayName += " " + row.LastName.String
			}
		}

		admins[i] = model.ChatAdmin{
			UserID:      row.ID,
			DisplayName: displayName,
			CreatedAt:   row.CreatedAt.Time,
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
