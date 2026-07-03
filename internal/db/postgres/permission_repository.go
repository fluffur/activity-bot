package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/permission"
	"context"
)

type PermissionRepository struct {
	queries *db.Queries
}

func NewPermissionRepository(queries *db.Queries) *PermissionRepository {
	return &PermissionRepository{queries: queries}
}

func (r *PermissionRepository) CommandPermission(ctx context.Context, chatID int64, name string) (permission.Status, error) {
	p, err := r.queries.GetCommandPermission(ctx, db.GetCommandPermissionParams{
		ChatID:     chatID,
		CommandKey: name,
	})
	if err != nil {
		return 0, err
	}

	return permission.Status(p.RequiredStatus), nil
}
