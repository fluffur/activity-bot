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

func (r *AdminRepository) GetChatsWithoutTitle(ctx context.Context) ([]model.Chat, error) {
	chats, err := r.queries.GetChatsWithoutTitle(ctx)
	if err != nil {
		return nil, err
	}
	return mapChats(chats), nil
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

func (r *AdminRepository) GetChatsWhereUserIsAdmin(ctx context.Context, userID int64) ([]model.Chat, error) {
	chats, err := r.queries.GetChatsWhereUserIsAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}

	mapped := make([]model.Chat, len(chats))
	for i, c := range chats {
		mapped[i] = mapChat(db.EnsureChatExistsRow(c))
	}

	return mapped, nil
}

func (r *AdminRepository) GetAllChats(ctx context.Context) ([]model.Chat, error) {
	chats, err := r.queries.GetAllChats(ctx)
	if err != nil {
		return nil, err
	}
	mapped := make([]model.Chat, len(chats))
	for i, c := range chats {
		mapped[i] = mapChat(db.EnsureChatExistsRow(c))
	}

	return mapped, nil
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
			ID: warn.ID,
			Moderator: model.User{
				ID:        warn.ModeratorID,
				FirstName: warn.FirstName.String,
				LastName:  warn.LastName.String,
				Username:  &warn.Username.String,
			},
			Reason:    warn.Reason.String,
			CreatedAt: warn.CreatedAt.Time,
			ExpiresAt: warn.ExpiresAt.Time,
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

func (r *AdminRepository) GetDeveloperRole(ctx context.Context, chatID, userID int64) (string, error) {
	dev, err := r.queries.GetDeveloper(ctx, db.GetDeveloperParams{
		UserID: userID,
		ChatID: chatID,
	})
	if err != nil {
		return "", err
	}
	return dev.Role, nil
}

func (r *AdminRepository) SetDeveloperRole(ctx context.Context, chatID, userID int64, role string) error {
	return r.queries.SetDeveloper(ctx, db.SetDeveloperParams{
		UserID: userID,
		ChatID: chatID,
		Role:   role,
	})
}

func (r *AdminRepository) RemoveDeveloperRole(ctx context.Context, chatID, userID int64) error {
	return r.queries.RemoveDeveloper(ctx, db.RemoveDeveloperParams{
		UserID: userID,
		ChatID: chatID,
	})
}

func (r *AdminRepository) IsDeveloper(ctx context.Context, chatID, userID int64) (bool, error) {
	return r.queries.IsDeveloper(ctx, db.IsDeveloperParams{
		UserID: userID,
		ChatID: chatID,
	})
}

func (r *AdminRepository) GetAllDevelopers(ctx context.Context, chatID int64) ([]model.User, []string, error) {
	rows, err := r.queries.GetAllDevelopers(ctx, chatID)
	if err != nil {
		return nil, nil, err
	}

	users := make([]model.User, len(rows))
	roles := make([]string, len(rows))
	for i, row := range rows {
		var username *string
		if row.Username.Valid {
			username = &row.Username.String
		}
		users[i] = model.User{
			ID:        row.UserID,
			FirstName: row.FirstName.String,
			LastName:  row.LastName.String,
			Username:  username,
		}
		roles[i] = row.Role
	}
	return users, roles, nil
}

func (r *AdminRepository) GetActiveWarnsByChat(ctx context.Context, chatID int64) ([]model.Warn, error) {
	warns, err := r.queries.GetActiveWarnsByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	results := make([]model.Warn, len(warns))
	for i, warn := range warns {
		results[i] = model.Warn{
			ID: warn.ID,
			User: model.User{
				ID:        warn.UserID,
				FirstName: warn.UserFirstName.String,
				LastName:  warn.UserLastName.String,
				Username:  &warn.UserUsername.String,
				Gender:    warn.UserGender,
			},
			Moderator: model.User{
				ID:        warn.ModeratorID,
				FirstName: warn.ModFirstName.String,
				LastName:  warn.ModLastName.String,
				Username:  &warn.ModUsername.String,
				Gender:    warn.ModGender,
			},
			Reason:    warn.Reason.String,
			CreatedAt: warn.CreatedAt.Time,
			ExpiresAt: warn.ExpiresAt.Time,
		}
	}
	return results, nil
}
