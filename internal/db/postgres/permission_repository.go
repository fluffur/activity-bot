package postgres

import (
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/postgres/sqlc"
	"context"
)

type PermissionRepository struct {
	queries *db.Queries
}

func NewPermissionRepository(queries *db.Queries) *PermissionRepository {
	return &PermissionRepository{queries: queries}
}

func (r *PermissionRepository) CommandPermission(ctx context.Context, chatID int64, name string) (chatmember.Status, error) {
	p, err := r.queries.GetCommandPermission(ctx, db.GetCommandPermissionParams{
		ChatID:     chatID,
		CommandKey: name,
	})
	if err != nil {
		return 0, err
	}

	return chatmember.Status(p.RequiredStatus), nil
}
