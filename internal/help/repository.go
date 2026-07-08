package help

import (
	"activity-bot/internal/permission"
	"context"
)

type PermissionRepository interface {
	CommandPermissions(ctx context.Context, chatID int64) (map[string]permission.Status, error)
}
