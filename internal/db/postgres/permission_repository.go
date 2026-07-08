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

func (r *PermissionRepository) CommandPermissions(ctx context.Context, chatID int64) (map[string]permission.Status, error) {
	rows, err := r.queries.GetCommandPermissions(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]permission.Status, len(rows))

	for _, row := range rows {
		result[row.CommandKey] = permission.Status(row.RequiredStatus)
	}

	return result, nil
}
